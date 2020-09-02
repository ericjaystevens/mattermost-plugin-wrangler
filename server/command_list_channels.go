package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/spf13/pflag"
)

const (
	flagTeamFilter    = "team-filter"
	flagChannelFilter = "channel-filter"
)

type listChannelsOptions struct {
	teamFilter    string
	channelFilter string
}

func getListChannelsFlagSet() *pflag.FlagSet {
	listChannelsFlagSet := pflag.NewFlagSet("list channels", pflag.ContinueOnError)
	listChannelsFlagSet.String(flagTeamFilter, "", "A filter value that team names must contain to be shown on the list")
	listChannelsFlagSet.String(flagChannelFilter, "", "A filter value that channel names must contain to be shown on the list")

	return listChannelsFlagSet
}

func parseListChannelsArgs(args []string) (listChannelsOptions, error) {
	var options listChannelsOptions

	listChannelsFlagSet := getListChannelsFlagSet()
	err := listChannelsFlagSet.Parse(args)
	if err != nil {
		return options, err
	}

	options.teamFilter, err = listChannelsFlagSet.GetString(flagTeamFilter)
	if err != nil {
		return options, err
	}

	options.channelFilter, err = listChannelsFlagSet.GetString(flagChannelFilter)
	if err != nil {
		return options, err
	}

	return options, nil
}

func (p *Plugin) runListChannelsCommand(values map[string]string, commandArgs interface{}) (string, error, error) {

	extra, ok := commandArgs.(*model.CommandArgs)
	if !ok {
		return "type mismatch error", fmt.Errorf("Expected type *model.CommandArgs actual: %s", reflect.TypeOf(commandArgs)), fmt.Errorf("Expected type *model.CommandArgs actual: %s", reflect.TypeOf(commandArgs))
	}

	teams, appErr := p.API.GetTeamsForUser(extra.UserId)
	if appErr != nil {
		return "", appErr, appErr
	}

	var msg string
	for _, team := range teams {
		if len(values["team-filter"]) != 0 && !strings.Contains(team.Name, values["team-filter"]) {
			continue
		}

		channels, appErr := p.API.GetChannelsForTeamForUser(team.Id, extra.UserId, false)
		if appErr != nil {
			return "", appErr, appErr
		}

		var filteredChannels []*model.Channel
		for _, channel := range channels {
			if channel.IsGroupOrDirect() {
				continue
			}
			if len(values["channel-filter"]) != 0 && !strings.Contains(channel.Name, values["channel-filter"]) {
				continue
			}
			filteredChannels = append(filteredChannels, channel)
		}
		if len(filteredChannels) == 0 {
			continue
		}

		// Format filtered channel list and append.
		newChannelGroup := fmt.Sprintf("%s\n", team.Name)
		for _, channel := range filteredChannels {
			newChannelGroup += fmt.Sprintf("%s - %s\n", channel.Id, channel.Name)
		}
		newChannelGroup = strings.TrimRight(newChannelGroup, "\n")
		msg += codeBlock(newChannelGroup) + "\n"
	}

	if len(msg) == 0 {
		msg = "No results found"
	}

	return msg, nil, nil
}
