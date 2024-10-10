package main

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

const (
	commandTriggerDialog = "parabol"

	dialogElementNameNumber = "somenumber"
	dialogElementNameEmail  = "someemail"

	dialogStateSome                = "somestate"
	dialogStateRelativeCallbackURL = "relativecallbackstate"

	commandDialogHelp = "###### Interactive Parabol Slash Command Help\n" +
		//"- `/dialog` - Open an Interactive Dialog. Once submitted, user-entered input is posted back into a channel.\n" +
		"- `/dialog start` - Start a Parabol Activity.\n" +//Open an Interactive Dialog with no elements. Once submitted, user's action is posted back into a channel.\n" +
		/*"- `/dialog relative-callback-url` - Open an Interactive Dialog with relative callback URL. Once submitted, user's action is posted back into a channel.\n" +
		"- `/dialog introduction-text` - Open an Interactive Dialog with optional introduction text. Once submitted, user's action is posted back into a channel.\n" +
		"- `/dialog error` - Open an Interactive Dialog which always returns an general error.\n" +
		"- `/dialog error-no-elements` - Open an Interactive Dialog with no elements which always returns an general error.\n" +
		*/
		"- `/dialog help` - Show this help text"
)

func (p *Plugin) registerCommands() error {
	if err := p.API.RegisterCommand(&model.Command{
		Trigger:          commandTriggerDialog,
		AutoComplete:     true,
		AutoCompleteDesc: "Start a Parabol Activity.",
		DisplayName:      "Parabol",
		AutocompleteData: getCommandDialogAutocompleteData(),
	}); err != nil {
		return errors.Wrapf(err, "failed to register %s command", commandTriggerDialog)
	}

	return nil
}

func getCommandDialogAutocompleteData() *model.AutocompleteData {
	command := model.NewAutocompleteData(commandTriggerDialog, "", "Start a Parabol Activity")

	command.AddCommand(model.NewAutocompleteData("start", "", "Start a Parabol Activity"))

	/*
	relativeCallbackURL := model.NewAutocompleteData("relative-callback-url", "", "Open an Interactive Dialog with a relative callback url.")
	command.AddCommand(relativeCallbackURL)

	introText := model.NewAutocompleteData("introduction-text", "", "Open an Interactive Dialog with an introduction text.")
	command.AddCommand(introText)

	error := model.NewAutocompleteData("error", "", "Open an Interactive Dialog with error.")
	command.AddCommand(error)

	errorNoElements := model.NewAutocompleteData("error-no-elements", "", "Open an Interactive Dialog with error no elements.")
	command.AddCommand(errorNoElements)
	*/

	command.AddCommand(model.NewAutocompleteData("help", "", ""))

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
	config := p.getConfiguration()
	url := config.ParabolURL
	var dialogRequest model.OpenDialogRequest
	fields := strings.Fields(args.Command)
	command := ""
	if len(fields) == 2 {
		command = fields[1]
	}


	user, _ := p.API.GetUser(args.UserId)
	fmt.Print("GEORG User", user.Email, user.EmailVerified)
	fmt.Print("GEORG url", url)

	switch command {
	case "help":
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         commandDialogHelp,
		}
	case "start":
		data := map[string]interface{}{"foo": "bar"}
		p.API.PublishWebSocketEvent("open_start_activity_modal", data, &model.WebsocketBroadcast{UserId: args.UserId})
		/*
		dialogRequest = model.OpenDialogRequest{
			TriggerId: args.TriggerId,
			URL:       fmt.Sprintf("/plugins/%s/start", manifest.Id),
			Dialog:    p.getStartActivityDialog(user.Email),
		}
		*/
	case "notifications":
		p.API.KVSet("notifications", []byte(args.ChannelId))
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "Enabled notifications",
		}
	case "dialog":
		dialogRequest = model.OpenDialogRequest{
			TriggerId: args.TriggerId,
			URL:       fmt.Sprintf("/plugins/%s/dialog/1", manifest.Id),
			Dialog:    getDialogWithIntroductionText("**Some** _introductory_ paragraph in Markdown formatted text with [link](https://mattermost.com)"),
		}
	case "error":
		data := map[string]interface{}{"error": "GEORG was here occurred"}

		p.API.PublishWebSocketEvent("error", data, &model.WebsocketBroadcast{UserId: args.UserId})
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

func (p *Plugin) getStartActivityDialog(email string) model.Dialog {
	templates := p.queryMeetingTemplates(email)
	if templates == nil {
		fmt.Print("GEORG error retrieving templates")
	}
	//state, _ := json.Marshal(templates)

	iconURL, _ := p.API.GetFileLink("parabol.svg")
	fmt.Print("GEORG iconURL", iconURL)

	var teams []*model.PostActionOptions
	for _, team := range templates.Teams {
		teams = append(teams, &model.PostActionOptions{
			Text:  team.Name,
			Value: team.ID,
		})
	}
	defaultTeam := ""
	if len(teams) > 0 {
		defaultTeam = teams[0].Value
	}

	var activities []*model.PostActionOptions
	for _, activity := range templates.AvailableTemplates {
		activities = append(activities, &model.PostActionOptions{
			Text:  activity.Name,
			Value: fmt.Sprintf("%s:%s", activity.Type, activity.ID),
		})
	}
	defaultActivity := ""
	if len(activities) > 0 {
		defaultActivity = activities[0].Value
	}

	return model.Dialog{
		CallbackId: "startActivity",
		Title:      "Start a Parabol Activity",
		IconURL:    iconURL,
		IntroductionText: "To see the full details for any activity, visit [Parabol's Activity Library](https://mattermost.com)",
		//State: string(state),
		Elements: []model.DialogElement{
			{
				DisplayName: "Choose Parabol Team",
				Name:        "team",
				Type:        "select",
				Placeholder: "Select a Team...",
				Default:     defaultTeam,
				Options:     teams,
			}, {
				DisplayName: "Choose Activity",
				Name:        "template",
				Type:        "select",
				Placeholder: "Select an Activity...",
				Default:     defaultActivity,
				Options:     activities,
			}},
		SubmitLabel:    "Start Activity",
		NotifyOnCancel: true,
	}
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
