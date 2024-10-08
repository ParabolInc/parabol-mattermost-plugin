package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	post, err := p.API.CreatePost(&model.Post{
		ChannelId: string(channel),
		Message:   "Hello, this is a notification from your plugin!",
		UserId:    string(userId),
	})
	fmt.Print("GEORG", post, err)
}


func (p *Plugin) templates(c *Context, w http.ResponseWriter, r *http.Request) {
	templates := p.queryMeetingTemplates(c.User.Email)
	w.Header().Set("Content-Type", "application/json")
        body, _ := json.Marshal(templates)
	w.Write(body)
}

func (p *Plugin) templates2(c *Context, w http.ResponseWriter, r *http.Request) {
	fmt.Print("GEORG GET TEMPLATES")
	p.templates(c, w, r)
}

func (p *Plugin) updateMeetingSettings(c *Context, w http.ResponseWriter, r *http.Request) {
	var settings SetMeetingSettingsVariables
	err := json.NewDecoder(r.Body).Decode(&settings)
	if err != nil {
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
	mux.HandleFunc("/templates", p.authenticated(p.templates))
	mux.HandleFunc("/templates2", p.authenticated(p.templates2))
	mux.HandleFunc("/meeting-settings", p.authenticated(p.updateMeetingSettings))
	mux.ServeHTTP(w, r)
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
