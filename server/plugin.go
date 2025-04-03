package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"

	"github.com/yaronf/httpsign"
)

const (
	botUserID      = "botUserID"
	requestTimeout = 30 * time.Second
	// well below the 4kb limit of nginx
	maxHeaderLength = 1024
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
}

type HTTPHandlerFuncWithContext func(c *Context, w http.ResponseWriter, r *http.Request)

func safeCopyHeader(from http.Header, header string, to http.Header) error {
	for _, value := range from.Values(header) {
		if len(header)+len(value) > maxHeaderLength {
			return errors.New("header too long")
		}
		// Prevent header injection by disallowing newline characters
		if strings.ContainsAny(header, "\r\n") {
			return errors.New("invalid characters in header")
		}
		to.Add(header, value)
	}
	return nil
}

func (p *Plugin) createContext(userID string) (*Context, context.CancelFunc) {
	user, _ := p.API.GetUser(userID)
	// TODO check email and email verified

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)

	context := &Context{
		Ctx:    ctx,
		UserID: userID,
		User:   user,
	}

	return context, cancel
}

func (p *Plugin) authenticated(handler HTTPHandlerFuncWithContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("Mattermost-User-ID")
		if userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error": "Not authorized"}`))
			return
		}

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
	// We don't have 10.1 deployed yet
	// pluginID := p.API.GetPluginID()
	pluginID := "co.parabol.action"
	path := "/plugins/" + pluginID
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = path + r.URL.Path
		handler(w, r)
	}
}

func (p *Plugin) notify(w http.ResponseWriter, r *http.Request) {
	config := p.getConfiguration()
	privKey := []byte(config.ParabolToken)
	verifier, err := NewVerifier(privKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Verify config error"}`))
		return
	}
	if err1 := httpsign.VerifyRequest("parabol", *verifier, r); err1 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Verification error"}`))
		return
	}

	channelID := r.PathValue("channelID")
	userID, err1 := p.API.KVGet(botUserID)
	if err1 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Bot User not found", "originalError": "%v"}`, err1)
		_, _ = w.Write([]byte(msg))
		return
	}

	var props map[string]interface{}
	if err := getJSON(r.Body, &props); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Error parsing body", "originalError": "%v"}`, err)
		_, _ = w.Write([]byte(msg))
		return
	}
	if _, err := p.API.CreatePost(&model.Post{
		ChannelId: channelID,
		Props:     props,
		UserId:    string(userID),
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Error posting notification", "originalError": "%v"}`, err)
		_, _ = w.Write([]byte(msg))
		return
	}
}

func (p *Plugin) login(c *Context, w http.ResponseWriter, r *http.Request) {
	var variables json.RawMessage
	if err := getJSON(r.Body, &variables); err != nil && err != io.EOF {
		w.WriteHeader(http.StatusBadRequest)
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
	res, err := client.Post(url, "application/json", bufio.NewReader(bytes.NewReader(requestBody)))
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

func (p *Plugin) graphql(c *Context, w http.ResponseWriter, r *http.Request) {
	config := p.getConfiguration()
	url := config.ParabolURL + "/graphql"
	privKey := []byte(config.ParabolToken)

	client, err := NewSigningClient(privKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "Signing error"}`))
		return
	}
	defer r.Body.Close()
	req, err1 := http.NewRequest("POST", url, r.Body)
	if err1 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Request error", "originalError": "%v"}`, err1)
		_, _ = w.Write([]byte(msg))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if errCopy := safeCopyHeader(r.Header, "x-application-authorization", req.Header); errCopy != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Header error", "originalError": "%v"}`, errCopy)
		_, _ = w.Write([]byte(msg))
		return
	}

	res, err2 := client.Do(req)
	if err2 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Request error", "originalError": "%v"}`, err2)
		_, _ = w.Write([]byte(msg))
		return
	}
	defer res.Body.Close()

	w.WriteHeader(res.StatusCode)
	_, _ = io.Copy(w, res.Body)
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
	_, _ = w.Write(body)
}

/*
Endpoint for module federation, serves components from parabol.
We cannot contact the Parabol instance directly from the webapp because of security settings on it.
*/
func (p *Plugin) components(w http.ResponseWriter, r *http.Request) {
	file := r.PathValue("file")
	config := p.getConfiguration()
	url := config.ParabolURL + "/components/" + file

	client := &http.Client{}
	res, err := client.Get(url)
	if err != nil {
		http.Error(w, "Server Error", http.StatusBadGateway)
		msg := fmt.Sprintf(`{"error": "Request error", "originalError": "%v"}`, err)
		_, _ = w.Write([]byte(msg))
		return
	}
	defer res.Body.Close()

	for header := range res.Header {
		if err := safeCopyHeader(res.Header, header, w.Header()); err != nil {
			http.Error(w, "Server Error", http.StatusInternalServerError)
			msg := fmt.Sprintf(`{"error": "Header error", "originalError": "%v"}`, err)
			_, _ = w.Write([]byte(msg))
			return
		}
	}

	w.WriteHeader(res.StatusCode)
	_, _ = io.Copy(w, res.Body)
}

func (p *Plugin) parabolRedirect(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")
	config := p.getConfiguration()
	url := config.ParabolURL + "/" + path
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func commandsEqual(a, b []SlashCommand) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Trigger != b[i].Trigger || a[i].Description != b[i].Description {
			return false
		}
	}
	return true
}

func (p *Plugin) connect(c* Context, w http.ResponseWriter, r *http.Request) {
	var config ClientConfig
	if err := getJSON(r.Body, &config); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		msg := fmt.Sprintf(`{"error": "Error parsing commands", "originalError": "%v"}`, err)
		_, _ = w.Write([]byte(msg))
		return
	}
	if !commandsEqual(p.commands, config.Commands) {
		p.commands = config.Commands
		if err := p.registerCommands(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			msg := fmt.Sprintf(`{"error": "Error registering commands", "originalError": "%v"}`, err)
			_, _ = w.Write([]byte(msg))
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /notify/{channelID}", p.fixedPath(p.notify))
	mux.HandleFunc("POST /login", p.authenticated(p.login))
	mux.HandleFunc("POST /graphql", p.authenticated(p.graphql))
	mux.HandleFunc("POST /connect", p.authenticated(p.connect))
	mux.HandleFunc("GET /config", p.authenticated(p.getConfig))
	mux.HandleFunc("GET /components/{file}", p.components)
	mux.HandleFunc("/parabol/{path...}", p.parabolRedirect)
	mux.ServeHTTP(w, r)
}
