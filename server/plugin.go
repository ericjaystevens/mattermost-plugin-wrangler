package main

import (
	"io/ioutil"
	"sync"

	"github.com/pkg/errors"

	"github.com/ericjaystevens/slashparse"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

const configPath = "/home/ec2-user/code/mattermost-plugin-wrangler/wrangler.yaml"

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	BotUserID string

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
	slashCommand  slashparse.SlashCommand
}

// BuildHash is the full git hash of the build.
var BuildHash string

// BuildHashShort is the short git hash of the build.
var BuildHashShort string

// BuildDate is the build date of the build.
var BuildDate string

// OnActivate runs when the plugin activates and ensures the plugin is properly
// configured.
func (p *Plugin) OnActivate() error {
	config := p.getConfiguration()
	err := config.IsValid()
	if err != nil {
		return errors.Wrap(err, "invalid config")
	}

	bot := &model.Bot{
		Username:    "wrangler",
		DisplayName: "Wrangler",
		Description: "Created by the Wrangler plugin.",
	}
	options := []plugin.EnsureBotOption{
		plugin.ProfileImagePath("assets/profile.png"),
	}

	botID, err := p.Helpers.EnsureBot(bot, options...)
	if err != nil {
		return errors.Wrap(err, "failed to ensure Wrangler bot")
	}
	p.BotUserID = botID

	slashDef, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	p.slashCommand, err = slashparse.NewSlashCommand(slashDef)
	if err != nil {
		return err
	}

	p.slashCommand.SetHandler("wrangler move thread", p.runMoveThreadCommand)
	p.slashCommand.SetHandler("wrangler copy thread", p.runCopyThreadCommand)
	p.slashCommand.SetHandler("wrangler attach message", p.runAttachMessageCommand)
	return p.API.RegisterCommand(getCommand(config.CommandAutoCompleteEnable))
}
