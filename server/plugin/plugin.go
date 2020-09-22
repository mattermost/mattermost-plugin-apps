// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package plugin

import (
	"net/http"

	"github.com/gorilla/mux"
	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/command"
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	myhttp "github.com/mattermost/mattermost-plugin-apps/server/http"
	"github.com/mattermost/mattermost-plugin-apps/server/http/api"
	"github.com/mattermost/mattermost-plugin-apps/server/http/dialog"
	"github.com/mattermost/mattermost-plugin-apps/server/http/helloapp"
)

type Configurator interface {
	Get() configurator.Config
	Refresh() error
	Store(*configurator.StoredConfig)
}

type Plugin struct {
	plugin.MattermostPlugin
	*configurator.BuildConfig
	mattermost *pluginapi.Client

	apps         *apps.Service
	command      command.Service
	configurator configurator.Service
	http         myhttp.Service
}

func NewPlugin(buildConfig *configurator.BuildConfig) *Plugin {
	return &Plugin{
		BuildConfig: buildConfig,
	}
}

func (p *Plugin) OnActivate() error {
	p.mattermost = pluginapi.NewClient(p.API)

	botUserID, err := p.mattermost.Bot.EnsureBot(&model.Bot{
		Username:    constants.BotUserName,
		DisplayName: constants.BotDisplayName,
		Description: constants.BotDescription,
	}, pluginapi.ProfileImagePath("assets/profile.png"))
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot account")
	}

	p.configurator = configurator.NewConfigurator(p.mattermost, p.BuildConfig, botUserID)
	p.apps = apps.NewService(p.mattermost, p.configurator)

	p.http = myhttp.NewService(mux.NewRouter(), p.apps,
		dialog.Init,
		helloapp.Init,
		api.Init,
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

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, req *http.Request) {
	p.http.ServeHTTP(c, w, req)
}

func (p *Plugin) UserHasBeenCreated(pluginContext *plugin.Context, user *model.User) {
	p.apps.Proxy.OnUserHasBeenCreated(pluginContext, user)
}

func (p *Plugin) UserHasJoinedChannel(pluginContext *plugin.Context, channelMember *model.ChannelMember, actingUser *model.User) {
	p.apps.Proxy.OnUserJoinedChannel(pluginContext, channelMember, actingUser)
}

func (p *Plugin) UserHasLeftChannel(pluginContext *plugin.Context, channelMember *model.ChannelMember, actingUser *model.User) {
	p.apps.Proxy.OnUserLeftChannel(pluginContext, channelMember, actingUser)
}

func (p *Plugin) UserHasJoinedTeam(pluginContext *plugin.Context, teamMember *model.TeamMember, actingUser *model.User) {
	p.apps.Proxy.OnUserJoinedTeam(pluginContext, teamMember, actingUser)
}

func (p *Plugin) UserHasLeftTeam(pluginContext *plugin.Context, teamMember *model.TeamMember, actingUser *model.User) {
	p.apps.Proxy.OnUserLeftTeam(pluginContext, teamMember, actingUser)
}
