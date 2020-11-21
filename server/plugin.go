// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	gohttp "net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/apps/configurator"
	"github.com/mattermost/mattermost-plugin-apps/server/apps/impl"
	"github.com/mattermost/mattermost-plugin-apps/server/command"
	"github.com/mattermost/mattermost-plugin-apps/server/http"
	"github.com/mattermost/mattermost-plugin-apps/server/http/dialog"
	"github.com/mattermost/mattermost-plugin-apps/server/http/helloapp"
	"github.com/mattermost/mattermost-plugin-apps/server/http/restapi"
)

type Plugin struct {
	plugin.MattermostPlugin
	*apps.BuildConfig
	mattermost *pluginapi.Client

	apps    *apps.Service
	command command.Service
	conf    apps.Configurator
	http    http.Service
}

func NewPlugin(buildConfig *apps.BuildConfig) *Plugin {
	return &Plugin{
		BuildConfig: buildConfig,
	}
}

func (p *Plugin) OnActivate() error {
	p.mattermost = pluginapi.NewClient(p.API)

	botUserID, err := p.mattermost.Bot.EnsureBot(&model.Bot{
		Username:    apps.BotUsername,
		DisplayName: apps.BotDisplayName,
		Description: apps.BotDescription,
	}, pluginapi.ProfileImagePath("assets/profile.png"))
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot account")
	}

	p.conf = configurator.NewConfigurator(p.mattermost, p.BuildConfig, botUserID)
	p.apps = impl.NewService(p.mattermost, p.conf)

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
	if p.conf == nil {
		// pre-activate, nothing to do.
		return nil
	}

	stored := apps.StoredConfig{}
	_ = p.mattermost.Configuration.LoadPluginConfiguration(&stored)
	return p.conf.RefreshConfig(&stored)
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	resp, _ := p.command.ExecuteCommand(c, args)
	return resp, nil
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w gohttp.ResponseWriter, req *gohttp.Request) {
	p.http.ServeHTTP(c, w, req)
}

func (p *Plugin) UserHasBeenCreated(pluginContext *plugin.Context, user *model.User) {
	_ = p.apps.API.Notify(apps.NewUserContext(user), apps.SubjectUserCreated)
}

func (p *Plugin) UserHasJoinedChannel(pluginContext *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	_ = p.apps.API.Notify(apps.NewChannelMemberContext(cm, actingUser), apps.SubjectUserJoinedChannel)
}

func (p *Plugin) UserHasLeftChannel(pluginContext *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	_ = p.apps.API.Notify(apps.NewChannelMemberContext(cm, actingUser), apps.SubjectUserLeftChannel)
}

func (p *Plugin) UserHasJoinedTeam(pluginContext *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	_ = p.apps.API.Notify(apps.NewTeamMemberContext(tm, actingUser), apps.SubjectUserJoinedTeam)
}

func (p *Plugin) UserHasLeftTeam(pluginContext *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	_ = p.apps.API.Notify(apps.NewTeamMemberContext(tm, actingUser), apps.SubjectUserLeftTeam)
}

func (p *Plugin) MessageHasBeenPosted(pluginContext *plugin.Context, post *model.Post) {
	_ = p.apps.API.Notify(apps.NewPostContext(post), apps.SubjectPostCreated)
}

func (p *Plugin) ChannelHasBeenCreated(pluginContext *plugin.Context, ch *model.Channel) {
	_ = p.apps.API.Notify(apps.NewChannelContext(ch), apps.SubjectChannelCreated)
}
