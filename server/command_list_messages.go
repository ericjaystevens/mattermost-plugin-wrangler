package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	flagListMessagesCount = "count"
	minListMessagesCount  = 1
	maxListMessagesCount  = 100

	flagListMessagesTrimLength = "trim-length"
	minListMessagesTrimLength  = 10
	maxListMessagesTrimLength  = 500
)

type listMessagesOptions struct {
	count      int
	trimLength int
}

// TODO: These validators should be handled by slashparse
func validateCount(value string) (count int, err error) {
	count, err = strconv.Atoi(value)
	if err != nil {
		return
	}

	if (count < minListMessagesCount) || (count > maxListMessagesCount) {
		err = fmt.Errorf("count (%d) must be between %d and %d", count, minListMessagesCount, maxListMessagesCount)
	}

	return
}

func validateLength(value string) (length int, err error) {
	length, err = strconv.Atoi(value)
	if err != nil {
		return
	}

	if (length < minListMessagesTrimLength) || (length > maxListMessagesTrimLength) {
		err = fmt.Errorf("%s (%d) must be between %d and %d", flagListMessagesTrimLength, length, minListMessagesTrimLength, maxListMessagesTrimLength)
	}

	return
}

func (p *Plugin) runListMessagesCommand(values map[string]string, extra *model.CommandArgs) (*model.CommandResponse, bool, error) {

	count, err := validateCount(values["count"])
	if err != nil {
		return nil, true, err
	}

	length, err := validateLength(values["trim-length"])
	if err != nil {
		return nil, true, err
	}

	channelPosts, appErr := p.API.GetPostsForChannel(extra.ChannelId, 0, count)
	if appErr != nil {
		return nil, false, appErr
	}

	msg := fmt.Sprintf("The last %d messages in this channel:\n", count)
	for _, post := range channelPosts.ToSlice() {
		if post.IsSystemMessage() {
			msg += "[     system message     ] - <skipped>\n"
		} else {
			msg += fmt.Sprintf("%s - %s\n", post.Id, cleanAndTrimMessage(post.Message, length))
		}
	}

	msg = codeBlock(strings.TrimRight(msg, "\n"))

	return getCommandResponse(model.COMMAND_RESPONSE_TYPE_EPHEMERAL, msg), false, nil
}
