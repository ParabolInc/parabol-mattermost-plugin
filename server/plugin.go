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

func (p *Plugin) notify(w http.ResponseWriter, r *http.Request) {
	userId, err := p.API.KVGet(botUserID)
	channel, err2 := p.API.KVGet("notifications")
	if err != nil || err2 != nil {
		return
	}

	var props map[string]interface{}
	err3 := getJson(r.Body, &props)
	if err3 != nil {
		return
	}
	_, err = p.API.CreatePost(&model.Post{
		ChannelId: string(channel),
		Props:    props,
		UserId:    string(userId),
	})
}


func (p *Plugin) templates(c *Context, w http.ResponseWriter, r *http.Request) {
	templates := p.queryMeetingTemplates(c.User.Email)
	w.Header().Set("Content-Type", "application/json")
        body, _ := json.Marshal(templates)
	w.Write(body)
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
	client := NewSigningClient(privKey)

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
		return
	}
	res, err := client.Post(url, "application/json", bufio.NewReader(bytes.NewReader(requestBody)))
	if err != nil || res.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Print("GEORG parabol err", err, res.StatusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	defer res.Body.Close()
	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Print("GEORG serialize err", err)
	} else {
		w.Write(responseBody)
	}
}

func (p *Plugin) updateMeetingSettings(c *Context, w http.ResponseWriter, r *http.Request) {
	var settings SetMeetingSettingsVariables
	err := json.NewDecoder(r.Body).Decode(&settings)
	if err != nil {
		fmt.Print("GEORG updateMeetingSettings, err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	p.setMeetingSettings(c.User.Email, settings)
	w.WriteHeader(http.StatusOK)
}

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	mux.HandleFunc("/start", p.startActivity)
	mux.HandleFunc("/notify", p.notify)
	mux.HandleFunc("POST /query/{query}", p.authenticated(p.query))
	mux.HandleFunc("/templates", p.authenticated(p.templates))
	mux.HandleFunc("/meeting-settings", p.authenticated(p.updateMeetingSettings))
	mux.ServeHTTP(w, r)
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
