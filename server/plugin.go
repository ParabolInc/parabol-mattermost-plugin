package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi/experimental/bot/logger"

	"github.com/yaronf/httpsign"
)

const (
	botUserID      = "botUserID"
	requestTimeout = 30 * time.Second
)

type SlashCommand struct {
	Trigger     string `json:"trigger"`
	Description string `json:"description"`
}

type ClientConfig struct {
	Commands []SlashCommand `json:"commands"`
}

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	commands []SlashCommand
}

type Context struct {
	Ctx    context.Context
	UserID string
	User   *model.User
	Log    logger.Logger
}

type HTTPHandlerFuncWithContext func(c *Context, w http.ResponseWriter, r *http.Request)

func copyRequestHeaders(from *http.Request, to *http.Request) {
	for header, values := range from.Header {
		for _, value := range values {
			to.Header.Add(header, value)
		}
	}
	to.Header.Set("X-Forwarded-For", from.RemoteAddr)
}
func forwardResponse(from *http.Response, to http.ResponseWriter) {
	for header, values := range from.Header {
		for _, value := range values {
			to.Header().Add(header, value)
		}
	}
	to.WriteHeader(from.StatusCode)
	_, _ = io.Copy(to, from.Body)
}

func (p *Plugin) createContext(userID string) (*Context, context.CancelFunc) {
	fmt.Print("createContext", userID, p)
	user, _ := p.API.GetUser(userID)
	// TODO check email and email verified

	logger := logger.New(p.API).With(logger.LogContext{
		"userid": userID,
	})

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)

	context := &Context{
		Ctx:    ctx,
		UserID: userID,
		User:   user,
		Log:    logger,
	}

	return context, cancel
}

func (p *Plugin) authenticated(handler HTTPHandlerFuncWithContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("Mattermost-User-ID")
		if userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error": "Not authorized"}`))
		}

		fmt.Print("authenticated", userID)
		context, cancel := p.createContext(userID)
		defer cancel()
		handler(context, w, r)
	}
}

/*
Mattermost strips the plugin path prefix from the request before forwarding it to the plugin.
If we want to verify the path of the request, we need to add it back.
https://github.com/mattermost/mattermost/blob/751d84bf13aa63f4706843318e45e8ca8401eba5/server/channels/app/plugin_requests.go#L226
*/
func (p *Plugin) fixedPath(handler http.HandlerFunc) http.HandlerFunc {
	// pre 10.1
	// pluginID := "co.parabol.action"
	pluginID := p.API.GetPluginID()
	path := "/plugins/" + pluginID
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = path + r.URL.Path
		handler(w, r)
	}
}

func (p *Plugin) verifiedFromParabol(handler http.HandlerFunc) http.HandlerFunc {
	return p.fixedPath(func(w http.ResponseWriter, r *http.Request) {
		config := p.getConfiguration()
		privKey := []byte(config.ParabolToken)
		verifier, err := NewVerifier(privKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "Verify config error"}`))
			return
		}
		if err = httpsign.VerifyRequest("parabol", *verifier, r); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "Verification error"}`))
			return
		}
		handler(w, r)
	})
}

func (p *Plugin) notify(w http.ResponseWriter, r *http.Request) {
	teamID := r.PathValue("teamID")
	userID, err2 := p.API.KVGet(botUserID)

	var props map[string]interface{}
	err3 := getJSON(r.Body, &props)
	channels, err4 := p.getChannels(teamID)

	if (err2 != nil) || (err3 != nil) || (err4 != nil) {
		return
	}
	for _, channel := range channels {
		_, err := p.API.CreatePost(&model.Post{
			ChannelId: channel,
			Props:     props,
			UserId:    string(userID),
		})
		if err != nil {
			fmt.Println("Post err", err)
		}
	}
}

func (p *Plugin) graphqlWebhook(w http.ResponseWriter, r *http.Request) {
	var message struct {
		Id string `json:"connectionId"`
		Payload string `json:"payload"`
	}
	if err := getJSON(r.Body, &message); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg := fmt.Sprintf(`{"error": "Error parsing message", "originalError": "%v"}`, err)
		_, _ = w.Write([]byte(msg))
		return
	}
	userId, _, found := strings.Cut(message.Id, "/")
	if userId == "" || !found {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Invalid connectionId"}`))
		return
	}
	data := map[string]interface{}{"id": message.Id, "payload": message.Payload}
	p.API.PublishWebSocketEvent("graphql", data, &model.WebsocketBroadcast{UserId: userId})
	w.WriteHeader(http.StatusOK)
}

func (p *Plugin) login(c *Context, w http.ResponseWriter, clientRequest *http.Request) {
	var variables json.RawMessage
	if err := getJSON(clientRequest.Body, &variables); err != nil && err != io.EOF {
		w.WriteHeader(http.StatusBadRequest)
		msg := fmt.Sprintf(`{"error": "Error parsing variables", "originalError": "%v"}`, err)
		_, _ = w.Write([]byte(msg))
		return
	}

	config := p.getConfiguration()
	url := config.ParabolURL + "/mattermost"
	privKey := []byte(config.ParabolToken)
	client, err := NewSigningClient(privKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Signing error"}`))
		return
	}

	query := struct {
		Email string `json:"email"`
	}{
		Email: c.User.Email,
	}
	requestBody, err := json.Marshal(query)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Marshal error"}`))
		return
	}

	parabolRequest, err1 := http.NewRequest("POST", url, bufio.NewReader(bytes.NewReader(requestBody)))
	if err1 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Error creating request"}`))
		return
	}
	copyRequestHeaders(clientRequest, parabolRequest)

	res, err := client.Do(parabolRequest)
	if err != nil || res.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Parabol server error", "originalError": "%v", "statusCode": "%v"}`, err, res.StatusCode)
		_, _ = w.Write([]byte(msg))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	defer res.Body.Close()
	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Serialization error"}`))
		return
	}
	if _, err = w.Write(responseBody); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Response error"}`))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (p *Plugin) graphql(c *Context, clientWriter http.ResponseWriter, clientRequest *http.Request) {
	config := p.getConfiguration()
	url := config.ParabolURL + "/graphql"
	privKey := []byte(config.ParabolToken)

	client, err := NewSigningClient(privKey)
	if err != nil {
		clientWriter.WriteHeader(http.StatusInternalServerError)
		_, _ = clientWriter.Write([]byte(`{"error": "Signing error"}`))
		return
	}
	defer clientRequest.Body.Close()
	parabolRequest, err1 := http.NewRequest("POST", url, clientRequest.Body)
	if err1 != nil {
		clientWriter.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Request error", "originalError": "%v"}`, err1)
		_, _ = clientWriter.Write([]byte(msg))
		return
	}
	copyRequestHeaders(clientRequest, parabolRequest)

	parabolResponse, err2 := client.Do(parabolRequest)
	if err2 != nil {
		http.Error(clientWriter, "Server Error", http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Request error", "originalError": "%v"}`, err2)
		_, _ = clientWriter.Write([]byte(msg))
		return
	}
	defer parabolResponse.Body.Close()

	forwardResponse(parabolResponse, clientWriter)
}

func (p *Plugin) linkedTeams(c *Context, w http.ResponseWriter, r *http.Request) {
	channelID := r.PathValue("channelID")
	teams, err := p.getTeams(channelID)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Could not read linked teams"}`))
		return
	}

	body, err := json.Marshal(teams)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Marshal error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Response error"}`))
		return
	}
}

func (p *Plugin) linkTeam(c *Context, w http.ResponseWriter, r *http.Request) {
	channelID := r.PathValue("channelID")
	teamID := r.PathValue("teamID")
	err := p.linkTeamToChannel(channelID, teamID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Error linking team to channel", "originalError": "%v"}`, err)
		_, _ = w.Write([]byte(msg))
		return
	}
}

func (p *Plugin) unlinkTeam(c *Context, w http.ResponseWriter, r *http.Request) {
	channelID := r.PathValue("channelID")
	teamID := r.PathValue("teamID")
	err := p.unlinkTeamFromChannel(channelID, teamID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Error unlinking team from channel"}`))
		return
	}
}

func (p *Plugin) getConfig(c *Context, w http.ResponseWriter, r *http.Request) {
	config := p.getConfiguration()
	body, err := json.Marshal(struct {
		ParabolURL string `json:"parabolUrl"`
	}{
		ParabolURL: config.ParabolURL,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Marshal error"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Response error"}`))
		return
	}
}

/*
Endpoint for module federation, serves components from parabol.
We cannot contact the Parabol instance directly from the webapp because of security settings on it.
*/
func (p *Plugin) components(clientWriter http.ResponseWriter, clientRequest *http.Request) {
	file := clientRequest.PathValue("file")
	config := p.getConfiguration()
	url := config.ParabolURL + "/components/" + file

	client := &http.Client{}
	parabolResponse, err := client.Get(url)
	if err != nil {
		http.Error(clientWriter, "Server Error", http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Request error", "originalError": "%v"}`, err)
		_, _ = clientWriter.Write([]byte(msg))
		return
	}
	defer parabolResponse.Body.Close()

	forwardResponse(parabolResponse, clientWriter)
}

func (p *Plugin) parabolRedirect(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")
	config := p.getConfiguration()
	url := config.ParabolURL + "/" + path
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (p *Plugin) connect(w http.ResponseWriter, r *http.Request) {
	var config = ClientConfig{}
	if err := getJSON(r.Body, &config); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg := fmt.Sprintf(`{"error": "Error parsing commands", "originalError": "%v"}`, err)
		_, _ = w.Write([]byte(msg))
		return
	}

	commands := config.Commands
	if err := p.loadCommands(commands); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Error loading commands", "originalError": "%v"}`, err)
		_, _ = w.Write([]byte(msg))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /notify/{teamID}", p.verifiedFromParabol(p.notify))
	mux.HandleFunc("POST /login", p.authenticated(p.login))
	mux.HandleFunc("POST /graphql", p.authenticated(p.graphql))
	mux.HandleFunc("POST /graphqlWebhook", p.verifiedFromParabol(p.graphqlWebhook))
	mux.HandleFunc("GET /linkedTeams/{channelID}", p.authenticated(p.linkedTeams))
	mux.HandleFunc("POST /linkTeam/{channelID}/{teamID}", p.authenticated(p.linkTeam))
	mux.HandleFunc("POST /unlinkTeam/{channelID}/{teamID}", p.authenticated(p.unlinkTeam))
	mux.HandleFunc("GET /config", p.authenticated(p.getConfig))
	mux.HandleFunc("GET /components/{file}", p.components)
	mux.HandleFunc("/parabol/{path...}", p.parabolRedirect)
	mux.HandleFunc("/connect", p.connect)
	mux.ServeHTTP(w, r)
}
