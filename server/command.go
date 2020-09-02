package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

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

	msg, userError, handlerErr := p.slashCommand.Execute(args.Command, args)

	if handlerErr != nil {
		p.API.LogError(handlerErr.Error())

		if userError != nil {
			return getCommandResponse(model.COMMAND_RESPONSE_TYPE_EPHEMERAL, userError.Error()), nil
		}

		return getCommandResponse(model.COMMAND_RESPONSE_TYPE_EPHEMERAL, "unknown error see logs for details."), nil
	}

	return getCommandResponse(model.COMMAND_RESPONSE_TYPE_EPHEMERAL, msg), nil
}

func (p *Plugin) runInfoCommand(nada map[string]string, extra interface{}) (string, error, error) {
	resp := fmt.Sprintf("Wrangler plugin version: %s, "+
		"[%s](https://github.com/gabrieljackson/mattermost-plugin-wrangler/commit/%s), built %s\n\n",
		manifest.Version, BuildHashShort, BuildHash, BuildDate)

	return resp, nil, nil
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
