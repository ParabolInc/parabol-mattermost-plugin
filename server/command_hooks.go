package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dadrus/httpsig"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

const (
	commandTriggerDialog            = "parabol"

	dialogElementNameNumber = "somenumber"
	dialogElementNameEmail  = "someemail"

	dialogStateSome                = "somestate"
	dialogStateRelativeCallbackURL = "relativecallbackstate"
	dialogIntroductionText         = "**Some** _introductory_ paragraph in Markdown formatted text with [link](https://mattermost.com)"

	commandDialogHelp = "###### Interactive Parabol Slash Command Help\n" +
		"- `/dialog` - Open an Interactive Dialog. Once submitted, user-entered input is posted back into a channel.\n" +
		"- `/dialog no-elements` - Open an Interactive Dialog with no elements. Once submitted, user's action is posted back into a channel.\n" +
		"- `/dialog relative-callback-url` - Open an Interactive Dialog with relative callback URL. Once submitted, user's action is posted back into a channel.\n" +
		"- `/dialog introduction-text` - Open an Interactive Dialog with optional introduction text. Once submitted, user's action is posted back into a channel.\n" +
		"- `/dialog error` - Open an Interactive Dialog which always returns an general error.\n" +
		"- `/dialog error-no-elements` - Open an Interactive Dialog with no elements which always returns an general error.\n" +
		"- `/dialog help` - Show this help text"
)

func (p *Plugin) registerCommands() error {
	if err := p.API.RegisterCommand(&model.Command{
		Trigger:          commandTriggerDialog,
		AutoComplete:     true,
		AutoCompleteDesc: "Open an Interactive Dialog.",
		DisplayName:      "Demo Plugin Command",
		AutocompleteData: getCommandDialogAutocompleteData(),
	}); err != nil {
		return errors.Wrapf(err, "failed to register %s command", commandTriggerDialog)
	}

	return nil
}

func getCommandDialogAutocompleteData() *model.AutocompleteData {
	command := model.NewAutocompleteData(commandTriggerDialog, "", "Open an Interactive Dialog.")

	noElements := model.NewAutocompleteData("foo-no-elements", "", "Open an Interactive Dialog with no elements.")
	command.AddCommand(noElements)

	relativeCallbackURL := model.NewAutocompleteData("relative-callback-url", "", "Open an Interactive Dialog with a relative callback url.")
	command.AddCommand(relativeCallbackURL)

	introText := model.NewAutocompleteData("introduction-text", "", "Open an Interactive Dialog with an introduction text.")
	command.AddCommand(introText)

	error := model.NewAutocompleteData("error", "", "Open an Interactive Dialog with error.")
	command.AddCommand(error)

	errorNoElements := model.NewAutocompleteData("error-no-elements", "", "Open an Interactive Dialog with error no elements.")
	command.AddCommand(errorNoElements)

	help := model.NewAutocompleteData("help", "", "")
	command.AddCommand(help)

	return command
}

// ExecuteCommand executes a command that has been previously registered via the RegisterCommand
// API.
//
// This demo implementation responds to a /demo_plugin command, allowing the user to enable
// or disable the demo plugin's hooks functionality (but leave the command and webapp enabled).
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	trigger := strings.TrimPrefix(strings.Fields(args.Command)[0], "/")
	switch trigger {
	case commandTriggerDialog:
		return p.executeCommandDialog(args), nil

	default:
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Unknown command: " + args.Command),
		}, nil
	}
}

func (p *Plugin) executeCommandDialog(args *model.CommandArgs) *model.CommandResponse {
	serverConfig := p.API.GetConfig()

	privKey := `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAABFwAAAAdzc2gtcn
NhAAAAAwEAAQAAAQEAujudq2xPiEXY0eFd2O58NC66czOT/eZdlsI43teJj2Twp4Yiyepj
hckFYsAmh1HCi7KlK6JAADQJkzmis6qVD7MXANDOG606tj3KddOj8NwDaysdFneUWVDqnS
AszXtKDPNo8K4ot7oU6OR4ox7HswqwHtm5ZANvaKDjQcUocAn8m3aaLnTKlWZiTcyJbxEP
y+JL7ZdB/LePTugPKoDEZ9t67oA6ll0/WEBkoH+OPwPRya/FynRNcakZ6KKBzlkRhGYaRQ
iCQG476uVpDOC0SGAmv/fD8g/1lFHHoIRvn4MeecKd0xbnzz3kLW8q0gDvVbAc0EjFEB9+
FyvD/EGtdwAAA8grCgEiKwoBIgAAAAdzc2gtcnNhAAABAQC6O52rbE+IRdjR4V3Y7nw0Lr
pzM5P95l2Wwjje14mPZPCnhiLJ6mOFyQViwCaHUcKLsqUrokAANAmTOaKzqpUPsxcA0M4b
rTq2Pcp106Pw3ANrKx0Wd5RZUOqdICzNe0oM82jwrii3uhTo5HijHsezCrAe2blkA29ooO
NBxShwCfybdpoudMqVZmJNzIlvEQ/L4kvtl0H8t49O6A8qgMRn23rugDqWXT9YQGSgf44/
A9HJr8XKdE1xqRnoooHOWRGEZhpFCIJAbjvq5WkM4LRIYCa/98PyD/WUUceghG+fgx55wp
3TFufPPeQtbyrSAO9VsBzQSMUQH34XK8P8Qa13AAAAAwEAAQAAAQBY2Mw11ixzVO9F4gDF
17EFrC1jfH3kKZ0IqYw8NBP6hyuQoJvEPMBSOT8Kh6VZ9ZWc1BOcp4FlF25iAKMwl/cZUF
VvHC7YYWKbQwtt/xQ9eple7WipKU9q9QGZCJqXRXRkjVPJTy05ydrj6Ovs1mhrcHPpo/Gg
V0s1XVxOKmNKX4fHw3Q1VCe/9lbOoDfiJxrOcjPMsUIyP0v0y8LhCpGGzRfr/jguSOP500
kIzWH+RNaGbJxrqWJyeSk/g9Nwrm2+EOgFP1hEp2ocppsqvs45jy7KaXQil7QVoUUYsKbZ
sXGrg7ost8V5r+P2kb33V/E40BufMWaN7ZTof9zeJTLxAAAAgQDd6WV2tny6e3zrG7tzDW
u6pkVpHvG/HH1HAIP5oNTTTMQ4lvfPUoWBNsAXQPXICZC0EL+sB6SAN4zHCuUaCiou2Fw/
6hWrVwfgjXHx9IeJJm/JnGBYDk0jKc084kYVUAb6gfMMCM+IG1ALlx/Q/XpP/pKXJaT4YI
ipLye/u9p6zwAAAIEA80Ip2qRCVgK6ZdgX/FMj1gCGS3DBSRnAUXmGP3Jkq8ql6ukzl8iB
zBHqXANoJouKjYxfAlZq0TOMLkJ2Fr0vypmgAnKY8rMbZx1YAn8wCmzI49wK4mLbFl5vPj
DnwdBj7Vng5u9EIQabOr7XfDFLvQtTOWNdj751Nvi8TGpBc4MAAACBAMP8zNrZE1EFQFOa
WHNY4o6wciwncdgsEhiLpR4rDgIrFR3XTJGrzbWCHPUkbraeNhqJb7kNcga/rhe12kECox
uucNMTb5CRPWwMLuX+a4i6F1k2r4iXHgW7PKEqIeHITqMj3awUEPxis89aqnG3zFSfvALB
+1Lufkk7Db4Zm1f9AAAAEGdlb3JnQGdyYW5ueS5sYW4BAg==
-----END OPENSSH PRIVATE KEY-----`
	signer, err := httpsig.NewSigner(
		// specify a key
		httpsig.Key{KeyID: "key1", Key: privKey, Algorithm: httpsig.EcdsaP256Sha256},
		// specify the required options
		// duration for which the signature should be valid
		httpsig.WithTTL(5 * time.Second),
		// which components should be protected by a signature
		httpsig.WithComponents("@authority", "@method", "x-my-fancy-header"),
		// a tag for your specific application
		httpsig.WithTag("myapp"),
	)
	// error handling goes here

	// Create a request
	req, err := http.NewRequestWithContext(context.Background(), "GET", "http://host.docker.internal:3001/mattermost", nil)
	// error handling goes here

	// Sign the request
	header, err := signer.Sign(httpsig.MessageFromRequest(req))
	// error handling goes here
	client := &http.Client{}

	// Add the signature to the request
	req.Header = header

	resp, err := client.Do(req)

	if err != nil {
		//log.Fatalln(err)
		fmt.Print("GEORG", err)
	} else {
	        fmt.Print("GEORG", resp)
	}

	var dialogRequest model.OpenDialogRequest
	fields := strings.Fields(args.Command)
	command := ""
	if len(fields) == 2 {
		command = fields[1]
	}

	switch command {
	case "help":
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         commandDialogHelp,
		}
	case "":
		dialogRequest = model.OpenDialogRequest{
			TriggerId: args.TriggerId,
			URL:       fmt.Sprintf("%s/plugins/%s/dialog/1", *serverConfig.ServiceSettings.SiteURL, manifest.Id),
			Dialog:    getDialogWithSampleElements(),
		}
	case "no-elements":
		dialogRequest = model.OpenDialogRequest{
			TriggerId: args.TriggerId,
			URL:       fmt.Sprintf("%s/plugins/%s/dialog/2", *serverConfig.ServiceSettings.SiteURL, manifest.Id),
			Dialog:    getDialogWithoutElements(dialogStateSome),
		}
	case "relative-callback-url":
		dialogRequest = model.OpenDialogRequest{
			TriggerId: args.TriggerId,
			URL:       fmt.Sprintf("/plugins/%s/dialog/2", manifest.Id),
			Dialog:    getDialogWithoutElements(dialogStateRelativeCallbackURL),
		}
	case "introduction-text":
		dialogRequest = model.OpenDialogRequest{
			TriggerId: args.TriggerId,
			URL:       fmt.Sprintf("%s/plugins/%s/dialog/1", *serverConfig.ServiceSettings.SiteURL, manifest.Id),
			Dialog:    getDialogWithIntroductionText(dialogIntroductionText),
		}
	case "error":
		dialogRequest = model.OpenDialogRequest{
			TriggerId: args.TriggerId,
			URL:       fmt.Sprintf("/plugins/%s/dialog/error", manifest.Id),
			Dialog:    getDialogWithSampleElements(),
		}
	case "error-no-elements":
		dialogRequest = model.OpenDialogRequest{
			TriggerId: args.TriggerId,
			URL:       fmt.Sprintf("/plugins/%s/dialog/error", manifest.Id),
			Dialog:    getDialogWithoutElements(dialogStateSome),
		}
	default:
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Unknown command: " + command),
		}
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
}

func getDialogWithSampleElements() model.Dialog {
	return model.Dialog{
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
			Name:        dialogElementNameEmail,
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
			Name:        dialogElementNameNumber,
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
		State:          dialogStateSome,
	}
}

func getDialogWithoutElements(state string) model.Dialog {
	return model.Dialog{
		CallbackId:     "somecallbackid",
		Title:          "Sample Confirmation Dialog",
		IconURL:        "http://www.mattermost.org/wp-content/uploads/2016/04/icon.png",
		Elements:       nil,
		SubmitLabel:    "Confirm",
		NotifyOnCancel: true,
		State:          state,
	}
}

func getDialogWithIntroductionText(introductionText string) model.Dialog {
	dialog := getDialogWithSampleElements()
	dialog.IntroductionText = introductionText
	return dialog
}

