// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package plugin

import (
	"net/http"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-cloudapps/server/cloudapps"
	"github.com/mattermost/mattermost-plugin-cloudapps/server/command"
	"github.com/mattermost/mattermost-plugin-cloudapps/server/configurator"
	"github.com/mattermost/mattermost-plugin-cloudapps/server/constants"
)

type Configurator interface {
	Get() configurator.Config
	Refresh() error
	Store(*configurator.StoredConfig)
}

type Plugin struct {
	plugin.MattermostPlugin
	*configurator.BuildConfig

	mm           *pluginapi.Client
	configurator configurator.Configurator
	command      command.Command
	registry     cloudapps.Registry
	proxy        cloudapps.Proxy

	botUserID string
}

func NewPlugin(buildConfig *configurator.BuildConfig) *Plugin {
	return &Plugin{
		BuildConfig: buildConfig,
	}
}

func (p *Plugin) OnActivate() error {
	p.mm = pluginapi.NewClient(p.API)

	botUserID, err := p.Helpers.EnsureBot(&model.Bot{
		Username:    constants.BotUserName,
		DisplayName: constants.BotDisplayName,
		Description: constants.BotDescription,
	}, plugin.ProfileImagePath("assets/profile.png"))
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot account")
	}
	p.botUserID = botUserID

	p.configurator = configurator.NewConfigurator(p.BuildConfig, p.mm)

	p.registry = cloudapps.NewRegistry(p.configurator)
	p.command = command.NewCommand(p.configurator, p.API)
	return p.command.Init(p.BuildConfig)
}

func (p *Plugin) OnConfigurationChange() error {
	if p.configurator == nil {
		// pre-activate, nothing to do.
		return nil
	}
	return p.configurator.Refresh()
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	out, err := p.command.Handle(c, args)
	if err != nil {
		return nil, model.NewAppError("Cloud Apps", "", nil, err.Error(), http.StatusInternalServerError)
	}
	p.mm.Post.SendEphemeralPost(args.UserId, &model.Post{
		ChannelId: args.ChannelId,
		UserId:    p.botUserID,
		Message:   out.String(),
	})
	return &model.CommandResponse{}, nil
}

func (p *Plugin) UserHasJoinedChannel(pluginContext *plugin.Context, channelMember *model.ChannelMember, actingUser *model.User) {
	p.proxy.OnUserJoinedChannel(pluginContext, channelMember, actingUser)
}
