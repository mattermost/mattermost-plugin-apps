// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package plugin

import (
	gohttp "net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/command"
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/http"
	"github.com/mattermost/mattermost-plugin-apps/server/http/dialog"
	"github.com/mattermost/mattermost-plugin-apps/server/http/helloapp"
	"github.com/mattermost/mattermost-plugin-apps/server/http/restapi"
)

type Plugin struct {
	plugin.MattermostPlugin
	*configurator.BuildConfig
	mattermost *pluginapi.Client

	apps         *apps.Service
	command      command.Service
	configurator configurator.Service
	http         http.Service
}

func NewPlugin(buildConfig *configurator.BuildConfig) *Plugin {
	return &Plugin{
		BuildConfig: buildConfig,
	}
}

func (p *Plugin) OnActivate() error {
	p.mattermost = pluginapi.NewClient(p.API)

	botUserID, err := p.mattermost.Bot.EnsureBot(&model.Bot{
		DisplayName: constants.BotDisplayName,
		Description: constants.BotDescription,
	}, pluginapi.ProfileImagePath("assets/profile.png"))
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot account")
	}

	p.configurator = configurator.NewConfigurator(p.mattermost, p.BuildConfig, botUserID)
	p.apps = apps.NewService(p.mattermost, p.configurator)

	p.http = http.NewService(mux.NewRouter(), p.apps,
		dialog.Init,
		helloapp.Init,
		restapi.Init,
	)

	p.command, err = command.MakeService(p.apps)
	if err != nil {
		return errors.Wrap(err, "failed to initialize own command handling")
	}
	return nil
}

func (p *Plugin) OnConfigurationChange() error {
	if p.configurator == nil {
		// pre-activate, nothing to do.
		return nil
	}
	return p.configurator.Refresh()
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	resp, _ := p.command.ExecuteCommand(c, args)
	return resp, nil
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w gohttp.ResponseWriter, req *gohttp.Request) {
	p.http.ServeHTTP(c, w, req)
}

func (p *Plugin) UserHasBeenCreated(pluginContext *plugin.Context, user *model.User) {
	_ = p.apps.Notify(constants.SubjectUserCreated, nil, nil, user, nil, nil)
}

func (p *Plugin) UserHasJoinedChannel(pluginContext *plugin.Context, channelMember *model.ChannelMember, actingUser *model.User) {
	_ = p.apps.Notify(constants.SubjectUserJoinedChannel, nil, channelMember, actingUser, nil, nil)
}

func (p *Plugin) UserHasLeftChannel(pluginContext *plugin.Context, channelMember *model.ChannelMember, actingUser *model.User) {
	_ = p.apps.Notify(constants.SubjectUserLeftChannel, nil, channelMember, actingUser, nil, nil)
}

func (p *Plugin) UserHasJoinedTeam(pluginContext *plugin.Context, teamMember *model.TeamMember, actingUser *model.User) {
	_ = p.apps.Notify(constants.SubjectUserJoinedTeam, teamMember, nil, actingUser, nil, nil)
}

func (p *Plugin) UserHasLeftTeam(pluginContext *plugin.Context, teamMember *model.TeamMember, actingUser *model.User) {
	_ = p.apps.Notify(constants.SubjectUserLeftTeam, teamMember, nil, actingUser, nil, nil)
}

func (p *Plugin) MessageHasBeenPosted(pluginContext *plugin.Context, post *model.Post) {
	_ = p.apps.Notify(constants.SubjectPostCreated, nil, nil, nil, nil, post)
}

func (p *Plugin) ChannelHasBeenCreated(pluginContext *plugin.Context, channel *model.Channel) {
	_ = p.apps.Notify(constants.SubjectChannelCreated, nil, nil, nil, channel, nil)
}
