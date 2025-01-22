package main

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

const (
	commandTrigger     = "parabol"
	commandDescription = "Start a Parabol Activity."
	commandHelpTitle   = "###### Parabol Slash Command Help"
)

func (p *Plugin) registerCommands() error {
	if err := p.API.RegisterCommand(&model.Command{
		Trigger:          commandTrigger,
		AutoComplete:     true,
		AutoCompleteDesc: commandDescription,
		DisplayName:      "Parabol",
		AutocompleteData: p.getCommandDialogAutocompleteData(),
	}); err != nil {
		return errors.Wrapf(err, "failed to register %s command", commandTrigger)
	}

	return nil
}

func (p *Plugin) getCommandDialogAutocompleteData() *model.AutocompleteData {
	command := model.NewAutocompleteData(commandTrigger, "", commandDescription)

	for _, commandDef := range p.commands {
		if commandDef.Trigger != "" {
			command.AddCommand(model.NewAutocompleteData(commandDef.Trigger, "", commandDef.Description))
		}
	}

	command.AddCommand(model.NewAutocompleteData("help", "", "Show help message"))

	return command
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	trigger := strings.TrimPrefix(strings.Fields(args.Command)[0], "/")
	switch trigger {
	case commandTrigger:
		return p.executeCommand(args), nil

	default:
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Unknown command: %s", args.Command),
		}, nil
	}
}

func (p *Plugin) executeCommand(args *model.CommandArgs) *model.CommandResponse {
	fields := strings.Fields(args.Command)
	command := ""
	if len(fields) >= 2 {
		command = fields[1]
	}

	switch command {
	case "help":
		helpText := commandHelpTitle
		if len(p.commands) == 0 {
			helpText += "\n\nFailed to connect to Parabol, check the configuration."
		} else {
			for _, commandDef := range p.commands {
				helpText += fmt.Sprintf("\n- `/%s %s` - %s", commandTrigger, commandDef.Trigger, commandDef.Description)
			}
		}

		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         helpText,
		}
	case "connect":
		if err := p.loadCommands(); err != nil {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         fmt.Sprintf("Failed to connect to Parabol, check the configuration (%s)", err),
			}
		}
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Successfully connected to Parabol",
		}
	// this case is left here for development, so it's easy to copy the styles
	case "dialog":
		dialogRequest := model.OpenDialogRequest{
			TriggerId: args.TriggerId,
			URL:       fmt.Sprintf("/plugins/%s/dialog/1", manifest.Id),
			Dialog:    getDialogWithSampleElements(),
		}
		if err := p.API.OpenInteractiveDialog(dialogRequest); err != nil {
			errorMessage := "Failed to open Interactive Dialog"
			p.API.LogError(errorMessage, "err", err.Error())
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         errorMessage,
			}
		}
		return &model.CommandResponse{}
	default:
		for _, commandDef := range p.commands {
			data := map[string]interface{}{"fields": fields}
			if commandDef.Trigger == command {
				p.API.PublishWebSocketEvent(commandDef.Trigger, data, &model.WebsocketBroadcast{UserId: args.UserId})
				return &model.CommandResponse{}
			}
		}

		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Unknown command: %s", command),
		}
	}
}

func getDialogWithSampleElements() model.Dialog {
	dialog := model.Dialog{
		CallbackId: "somecallbackid",
		Title:      "Test Title",
		IconURL:    "http://www.mattermost.org/wp-content/uploads/2016/04/icon.png",
		Elements: []model.DialogElement{{
			DisplayName: "Display Name",
			Name:        "realname",
			Type:        "text",
			Default:     "default text",
			Placeholder: "placeholder",
			HelpText:    "This a regular input in an interactive dialog triggered by a test integration.",
		}, {
			DisplayName: "Email",
			Name:        "someemail",
			Type:        "text",
			SubType:     "email",
			Placeholder: "placeholder@bladekick.com",
			HelpText:    "This a regular email input in an interactive dialog triggered by a test integration.",
		}, {
			DisplayName: "Password",
			Name:        "somepassword",
			Type:        "text",
			SubType:     "password",
			Placeholder: "Password",
			HelpText:    "This a password input in an interactive dialog triggered by a test integration.",
		}, {
			DisplayName: "Number",
			Name:        "somenumber",
			Type:        "text",
			SubType:     "number",
		}, {
			DisplayName: "Display Name Long Text Area",
			Name:        "realnametextarea",
			Type:        "textarea",
			Placeholder: "placeholder",
			Optional:    true,
			MinLength:   5,
			MaxLength:   100,
		}, {
			DisplayName: "User Selector",
			Name:        "someuserselector",
			Type:        "select",
			Placeholder: "Select a user...",
			HelpText:    "Choose a user from the list.",
			Optional:    true,
			MinLength:   5,
			MaxLength:   100,
			DataSource:  "users",
		}, {
			DisplayName: "Channel Selector",
			Name:        "somechannelselector",
			Type:        "select",
			Placeholder: "Select a channel...",
			HelpText:    "Choose a channel from the list.",
			Optional:    true,
			MinLength:   5,
			MaxLength:   100,
			DataSource:  "channels",
		}, {
			DisplayName: "Option Selector",
			Name:        "someoptionselector",
			Type:        "select",
			Placeholder: "Select an option...",
			HelpText:    "Choose an option from the list.",
			Options: []*model.PostActionOptions{{
				Text:  "Option1",
				Value: "opt1",
			}, {
				Text:  "Option2",
				Value: "opt2",
			}, {
				Text:  "Option3",
				Value: "opt3",
			}},
		}, {
			DisplayName: "Option Selector with default",
			Name:        "someoptionselector2",
			Type:        "select",
			Default:     "opt2",
			Placeholder: "Select an option...",
			HelpText:    "Choose an option from the list.",
			Options: []*model.PostActionOptions{{
				Text:  "Option1",
				Value: "opt1",
			}, {
				Text:  "Option2",
				Value: "opt2",
			}, {
				Text:  "Option3",
				Value: "opt3",
			}},
		}, {
			DisplayName: "Boolean Selector",
			Name:        "someboolean",
			Type:        "bool",
			Placeholder: "Agree to the terms of service",
			HelpText:    "You must agree to the terms of service to proceed.",
		}, {
			DisplayName: "Boolean Selector",
			Name:        "someboolean_optional",
			Type:        "bool",
			Placeholder: "Sign up for monthly emails?",
			HelpText:    "It's up to you if you want to get monthly emails.",
			Optional:    true,
		}, {
			DisplayName: "Boolean Selector (default true)",
			Name:        "someboolean_default_true",
			Type:        "bool",
			Placeholder: "Enable secure login",
			HelpText:    "You must enable secure login to proceed.",
			Default:     "true",
		}, {
			DisplayName: "Boolean Selector (default true)",
			Name:        "someboolean_default_true_optional",
			Type:        "bool",
			Placeholder: "Enable painfully secure login",
			HelpText:    "You may optionally enable painfully secure login.",
			Default:     "true",
			Optional:    true,
		}, {
			DisplayName: "Boolean Selector (default false)",
			Name:        "someboolean_default_false",
			Type:        "bool",
			Placeholder: "Agree to the annoying terms of service",
			HelpText:    "You must also agree to the annoying terms of service to proceed.",
			Default:     "false",
		}, {
			DisplayName: "Boolean Selector (default false)",
			Name:        "someboolean_default_false_optional",
			Type:        "bool",
			Placeholder: "Throw-away account",
			HelpText:    "A throw-away account will be deleted after 24 hours.",
			Default:     "false",
			Optional:    true,
		}, {
			DisplayName: "Radio Option Selector",
			Name:        "someradiooptionselector",
			Type:        "radio",
			HelpText:    "Choose an option from the list.",
			Options: []*model.PostActionOptions{{
				Text:  "Option1",
				Value: "opt1",
			}, {
				Text:  "Option2",
				Value: "opt2",
			}, {
				Text:  "Option3",
				Value: "opt3",
			}},
		}},
		SubmitLabel:    "Submit",
		NotifyOnCancel: true,
		State:          "somestate",
	}
	dialog.IntroductionText = "**Some** _introductory_ paragraph in Markdown formatted text with [link](https://mattermost.com)"
	return dialog
}
