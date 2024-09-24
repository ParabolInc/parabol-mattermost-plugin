package main

import (
	"fmt"

	"github.com/pkg/errors"
)

// OnActivate is invoked when the plugin is activated.
//
// This demo implementation logs a message to the demo channel whenever the plugin is activated.
// It also creates a demo bot account
func (p *Plugin) OnActivate() error {
	fmt.Println("GEORG config", p.API.GetConfig().ServiceSettings.SiteURL)
	fmt.Printf("GEORG /plugins/%s/start", manifest.Id)
	if err := p.registerCommands(); err != nil {
		return errors.Wrap(err, "failed to register commands")
	}

	return nil
}

// OnDeactivate is invoked when the plugin is deactivated. This is the plugin's last chance to use
// the API, and the plugin will be terminated shortly after this invocation.
//
// This demo implementation logs a message to the demo channel whenever the plugin is deactivated.
func (p *Plugin) OnDeactivate() error {
	return nil
}

