package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi/experimental/bot/logger"

	"github.com/yaronf/httpsign"
)

const (
	botUserID = "botUserID"
	requestTimeout      = 30 * time.Second
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

type Context struct {
	Ctx    context.Context
	UserID string
	User   *model.User
	Log    logger.Logger
}

type HTTPHandlerFuncWithContext func(c *Context, w http.ResponseWriter, r *http.Request)

func (p *Plugin) createContext(userID string) (*Context, context.CancelFunc) {
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
	w.Write([]byte(`{"error": "Not authorized"}`))
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
  // from 10.1 we can use p.API.GetPluginID()
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
		w.Write([]byte(`{"error": "Verify config error"}`))
		return
	}
	err = httpsign.VerifyRequest("parabol", *verifier, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Verification error"}`))
		return
	}

	teamId := r.PathValue("teamId")
	userId, err := p.API.KVGet(botUserID)

	var props map[string]interface{}
	err3 := getJson(r.Body, &props)
	if err3 != nil {
		return
	}
	channels, err2 := p.getChannels(teamId)
	if err2 != nil {
		return
	}
	for _, channel := range channels {
		_, err = p.API.CreatePost(&model.Post{
			ChannelId: channel,
			Props:    props,
			UserId:    string(userId),
		})
	}
}

func (p *Plugin) query(c *Context, w http.ResponseWriter, r *http.Request) {
	queryRequest := r.PathValue("query")

	var variables json.RawMessage
	err := getJson(r.Body, &variables)
	if err != nil && err != io.EOF {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	config := p.getConfiguration()
	url := config.ParabolURL
	privKey := []byte(config.ParabolToken)
	client, err := NewSigningClient(privKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Signing error"}`))
		return
	}

	query := struct {
		Query     string    `json:"query"`
		Variables json.RawMessage `json:"variables"`
		Email     string    `json:"email"`
	}{
		Query: queryRequest,
		Variables: variables,
		Email: c.User.Email,
	}
	requestBody, err := json.Marshal(query)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Marshal error"}`))
		return
	}
	res, err := client.Post(url, "application/json", bufio.NewReader(bytes.NewReader(requestBody)))
	if err != nil || res.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Parabol server error", "originalError": "%v", "statusCode": "%v"}`, err, res.StatusCode)
		w.Write([]byte(msg))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	defer res.Body.Close()
	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Serialization error"}`))
		return
	}
	w.Write(responseBody)
}

func (p *Plugin) linkedTeams(c *Context, w http.ResponseWriter, r *http.Request) {
	channelId := r.PathValue("channelId")
	teams, err := p.getTeams(channelId)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Could not read linked teams"}`))
		return
	}

	body, err := json.Marshal(teams)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Marshal error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (p *Plugin) linkTeam(c *Context, w http.ResponseWriter, r *http.Request) {
	channelId := r.PathValue("channelId")
	teamId := r.PathValue("teamId")
	err := p.linkTeamToChannel(channelId, teamId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf(`{"error": "Error linking team to channel", "originalError": "%v"}`, err)
		w.Write([]byte(msg))
		return
	}
}

func (p *Plugin) unlinkTeam(c *Context, w http.ResponseWriter, r *http.Request) {
	channelId := r.PathValue("channelId")
	teamId := r.PathValue("teamId")
	err := p.linkTeamToChannel(channelId, teamId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Error unlinking team from channel"}`))
		return
	}
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /notify/{teamId}", p.fixedPath(p.notify))
	mux.HandleFunc("POST /query/{query}", p.authenticated(p.query))
	mux.HandleFunc("/linkedTeams/{channelId}", p.authenticated(p.linkedTeams))
	mux.HandleFunc("POST /linkTeam/{channelId}/{teamId}", p.authenticated(p.linkTeam))
	mux.HandleFunc("POST /unlinkTeam/{channelId}/{teamId}", p.authenticated(p.unlinkTeam))
	mux.ServeHTTP(w, r)
}
