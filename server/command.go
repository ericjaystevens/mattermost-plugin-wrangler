package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

const helpText = `Wrangler Plugin - Slash Command Help

%s

%s

/wrangler attach message [MESSAGE_ID_TO_ATTACH] [ROOT_MESSAGE_ID]
  Attach a given message to a thread in the same channel
    - Obtain the message IDs by running '/wrangler list messages' or via the 'Permalink' message dropdown option (it's the last part of the URL)

/wrangler list channels [flags]
  List the IDs of all channels you have joined
	Flags:
%s
/wrangler list messages [flags]
  List the IDs of recent messages in this channel
    Flags:
/wrangler info
  Shows plugin information`

func getHelp() string {
	return codeBlock(fmt.Sprintf(
		helpText,
		getMoveThreadUsage(),
		copyThreadUsage,
		getListChannelsFlagSet().FlagUsages(),
	))
}

func getCommand(autocomplete bool) *model.Command {
	return &model.Command{
		Trigger:          "wrangler",
		DisplayName:      "Wrangler",
		Description:      "Manage Mattermost messages!",
		AutoComplete:     autocomplete,
		AutoCompleteDesc: "Available commands: move thread, copy thread, attach message, list messages, list channels, info",
		AutoCompleteHint: "[command]",
		AutocompleteData: getAutocompleteData(),
	}
}

func getCommandResponse(responseType, text string) *model.CommandResponse {
	return &model.CommandResponse{
		ResponseType: responseType,
		Text:         text,
		Username:     "wrangler",
		IconURL:      fmt.Sprintf("/plugins/%s/profile.png", manifest.Id),
	}
}

// ExecuteCommand executes a given command and returns a command response.
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	if !p.authorizedPluginUser(args.UserId) {
		return getCommandResponse(model.COMMAND_RESPONSE_TYPE_EPHEMERAL, "Permission denied. Please talk to your system administrator to get access."), nil
	}

	stringArgs := strings.Split(args.Command, " ")

	wranglerParser := p.slashCommand

	//this block goes away when all command strings are handled by Execute
	slashCommand, values, err := wranglerParser.Parse(args.Command)
	if err != nil {
		return getCommandResponse(model.COMMAND_RESPONSE_TYPE_EPHEMERAL, err.Error()), nil
	}

	var handler func([]string, *model.CommandArgs) (*model.CommandResponse, bool, error)

	var resp *model.CommandResponse
	var userError bool //should the user be presented with error
	var handlerErr error
	var msg string
	//var friendlyErr error

	//hopefully this switch statement can go away and slashCommand.Execute() can replace it.
	switch slashCommand {
	case "wrangler move thread", "wrangler copy thread", "wrangler attach message":
		msg, _, handlerErr = p.slashCommand.Execute(args.Command, args)
	case "wrangler list channels":
		resp, userError, handlerErr = p.runListChannelsCommand(values, args)
	case "wrangler list messages":
		resp, userError, handlerErr = p.runListMessagesCommand(values, args)
	case "wrangler info":
		handler = p.runInfoCommand
		stringArgs = stringArgs[2:]
	default:
		msg = getHelp()
	}

	if msg != "" {
		resp = model.CommandResponseFromPlainText(msg)
	}

	if handler != nil {
		resp, userError, err = handler(stringArgs, args)
	}

	if handlerErr != nil {
		err = handlerErr
	}

	if err != nil {
		p.API.LogError(err.Error())
		if userError {
			return getCommandResponse(model.COMMAND_RESPONSE_TYPE_EPHEMERAL, fmt.Sprintf("__Error: %s__\n\nRun `/wrangler help` for usage instructions.", err.Error())), nil
		}

		return getCommandResponse(model.COMMAND_RESPONSE_TYPE_EPHEMERAL, "An unknown error occurred. Please talk to your administrator for help."), nil
	}

	return resp, nil
}

func (p *Plugin) runInfoCommand(args []string, extra *model.CommandArgs) (*model.CommandResponse, bool, error) {
	resp := fmt.Sprintf("Wrangler plugin version: %s, "+
		"[%s](https://github.com/gabrieljackson/mattermost-plugin-wrangler/commit/%s), built %s\n\n",
		manifest.Version, BuildHashShort, BuildHash, BuildDate)

	return getCommandResponse(model.COMMAND_RESPONSE_TYPE_EPHEMERAL, resp), false, nil
}

func (p *Plugin) authorizedPluginUser(userID string) bool {
	config := p.getConfiguration()

	if len(config.AllowedEmailDomain) != 0 {
		user, err := p.API.GetUser(userID)
		if err != nil {
			return false
		}

		emailDomains := strings.Split(config.AllowedEmailDomain, ",")
		for _, emailDomain := range emailDomains {
			if strings.HasSuffix(user.Email, emailDomain) {
				return true
			}
		}

		return false
	}

	return true
}

func getAutocompleteData() *model.AutocompleteData {
	wrangler := model.NewAutocompleteData("wrangler", "[command]", "Available commands: move, copy, attach, list, info, help")

	move := model.NewAutocompleteData("move", "[subcommand]", "Move messages")
	moveThread := model.NewAutocompleteData("thread", "[MESSAGE_ID] [CHANNEL_ID]", "Move a message and the thread it belongs to")
	moveThread.AddTextArgument("The ID of the message to be moved", "[MESSAGE_ID]", "")
	moveThread.AddTextArgument("The ID of the channel where the message will be moved to", "[CHANNEL_ID]", "")
	move.AddCommand(moveThread)
	wrangler.AddCommand(move)

	copy := model.NewAutocompleteData("copy", "[subcommand]", "Copy messages")
	copyThread := model.NewAutocompleteData("thread", "[MESSAGE_ID] [CHANNEL_ID]", "Copy a message and the thread it belongs to")
	copyThread.AddTextArgument("The ID of the message to be copied", "[MESSAGE_ID]", "")
	copyThread.AddTextArgument("The ID of the channel where the message will be copied to", "[CHANNEL_ID]", "")
	copy.AddCommand(copyThread)
	wrangler.AddCommand(copy)

	attach := model.NewAutocompleteData("attach", "[subcommand]", "Attach messages")
	attachMessage := model.NewAutocompleteData("message", "[MESSAGE_ID_TO_ATTACH] [ROOT_MESSAGE_ID]", "Attach a message to a thread in the channel")
	attachMessage.AddTextArgument("The ID of the message to be attached", "[MESSAGE_ID_TO_ATTACH]", "")
	attachMessage.AddTextArgument("The root message ID of the thread", "[ROOT_MESSAGE_ID]", "")
	attach.AddCommand(attachMessage)
	wrangler.AddCommand(attach)

	list := model.NewAutocompleteData("list", "[subcommand]", "Lists IDs for channels and messages")
	listChannels := model.NewAutocompleteData("channels", "[optional flags]", "List channel IDs that you have joined")
	listMessages := model.NewAutocompleteData("messages", "[optional flags]", "List message IDs in this channel")
	list.AddCommand(listChannels)
	list.AddCommand(listMessages)
	wrangler.AddCommand(list)

	info := model.NewAutocompleteData("info", "", "Shows plugin information")
	wrangler.AddCommand(info)

	help := model.NewAutocompleteData("help", "", "Shows detailed help information")
	wrangler.AddCommand(help)

	return wrangler
}
