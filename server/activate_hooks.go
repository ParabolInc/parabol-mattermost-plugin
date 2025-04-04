package main

import (
	"os"
	"path/filepath"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pkg/errors"
)

// OnActivate is invoked when the plugin is activated.
//
// This demo implementation logs a message to the demo channel whenever the plugin is activated.
// It also creates a demo bot account
func (p *Plugin) OnActivate() error {
	p.commands = []SlashCommand{{
		Trigger:     "check",
		Description: "Check if the configured Parabol server is reachable",
	}}
	if err := p.registerCommands(); err != nil {
		return errors.Wrap(err, "failed to register commands")
	}

	botID, err := p.API.EnsureBotUser(&model.Bot{
		Username:    "parabol",
		DisplayName: "Parabol",
		Description: "Created by the Parabol plugin.",
	})
	if err != nil {
		return errors.Wrap(err, "failed to create bot")
	}
	{
		bundlePath, err := p.API.GetBundlePath()
		if err != nil {
			return errors.Wrap(err, "failed to get bundle path")
		}

		profileImage, err := os.ReadFile(filepath.Join(bundlePath, "assets", "parabol.png"))
		if err != nil {
			return errors.Wrap(err, "failed to read profile image")
		}

		if err := p.API.SetProfileImage(botID, profileImage); err != nil {
			return errors.Wrap(err, "failed to set profile image")
		}
	}

	err2 := p.API.KVSet(botUserID, []byte(botID))
	if err2 != nil {
		return errors.Wrap(err2, "failed to store bot user ID")
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
