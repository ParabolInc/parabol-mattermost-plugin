package main

import (
	"net/http"
	"sync"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
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


func (p *Plugin) notify(w http.ResponseWriter, r *http.Request) {
	p.API.CreatePost(&model.Post{
		ChannelId: "town-square",
		Message:   "Hello, this is a notification from your plugin!",
		UserId:    "parabol",
	})
}

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	mux.HandleFunc("/start", p.startActivity)
	mux.HandleFunc("/notify", p.notify)
	mux.ServeHTTP(w, r)
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
