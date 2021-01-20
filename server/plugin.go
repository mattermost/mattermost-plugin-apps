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

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/admin"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/aws"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/configurator"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/store"
	"github.com/mattermost/mattermost-plugin-apps/server/command"
	"github.com/mattermost/mattermost-plugin-apps/server/examples/go/hello/builtin_hello"
	"github.com/mattermost/mattermost-plugin-apps/server/examples/go/hello/http_hello"
	"github.com/mattermost/mattermost-plugin-apps/server/http"
	"github.com/mattermost/mattermost-plugin-apps/server/http/dialog"
	"github.com/mattermost/mattermost-plugin-apps/server/http/restapi"
)

type Plugin struct {
	plugin.MattermostPlugin
	*api.BuildConfig

	mm      *pluginapi.Client
	api     *api.Service
	command command.Service
	http    http.Service
}

func NewPlugin(buildConfig *api.BuildConfig) *Plugin {
	return &Plugin{
		BuildConfig: buildConfig,
	}
}

func (p *Plugin) OnActivate() error {
	mm := pluginapi.NewClient(p.API)
	p.mm = mm

	botUserID, err := mm.Bot.EnsureBot(&model.Bot{
		Username:    api.BotUsername,
		DisplayName: api.BotDisplayName,
		Description: api.BotDescription,
	}, pluginapi.ProfileImagePath("assets/profile.png"))
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot account")
	}

	stored := api.StoredConfig{}
	_ = p.mm.Configuration.LoadPluginConfiguration(&stored)

	awsClient := aws.NewAWSClient(stored.AWSAccessKeyID, stored.AWSSecretAccessKey, &mm.Log)

	conf := configurator.NewConfigurator(mm, awsClient, p.BuildConfig, botUserID)
	store := store.NewStore(mm, conf)
	proxy := proxy.NewProxy(mm, awsClient, conf, store)

	p.api = &api.Service{
		Mattermost:   mm,
		Configurator: conf,
		Proxy:        proxy,
		AppServices:  appservices.NewAppServices(mm, conf, store),
		Admin:        admin.NewAdmin(mm, conf, store, proxy, awsClient),
		AWS:          awsClient,
	}
	proxy.ProvisionBuiltIn(builtin_hello.AppID, builtin_hello.New(p.api))

	p.http = http.NewService(mux.NewRouter(), p.api,
		dialog.Init,
		restapi.Init,
		http_hello.Init,
	)

	p.command, err = command.MakeService(p.api)
	if err != nil {
		return errors.Wrap(err, "failed to initialize own command handling")
	}

	if err := p.api.Admin.SynchronizeApps(); err != nil {
		mm.Log.Error("Can't synchronize", "err", err.Error())
	}

	return nil
}

func (p *Plugin) OnConfigurationChange() error {
	if p.api == nil || p.api.Configurator == nil {
		// pre-activate, nothing to do.
		return nil
	}

	stored := api.StoredConfig{}
	_ = p.mm.Configuration.LoadPluginConfiguration(&stored)
	return p.api.Configurator.RefreshConfig(&stored)
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	resp, _ := p.command.ExecuteCommand(c, args)
	return resp, nil
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w gohttp.ResponseWriter, req *gohttp.Request) {
	p.http.ServeHTTP(c, w, req)
}

func (p *Plugin) UserHasBeenCreated(pluginContext *plugin.Context, user *model.User) {
	_ = p.api.Proxy.Notify(api.NewUserContext(user), api.SubjectUserCreated)
}

func (p *Plugin) UserHasJoinedChannel(pluginContext *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	_ = p.api.Proxy.Notify(api.NewChannelMemberContext(cm, actingUser), api.SubjectUserJoinedChannel)
}

func (p *Plugin) UserHasLeftChannel(pluginContext *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	_ = p.api.Proxy.Notify(api.NewChannelMemberContext(cm, actingUser), api.SubjectUserLeftChannel)
}

func (p *Plugin) UserHasJoinedTeam(pluginContext *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	_ = p.api.Proxy.Notify(api.NewTeamMemberContext(tm, actingUser), api.SubjectUserJoinedTeam)
}

func (p *Plugin) UserHasLeftTeam(pluginContext *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	_ = p.api.Proxy.Notify(api.NewTeamMemberContext(tm, actingUser), api.SubjectUserLeftTeam)
}

func (p *Plugin) MessageHasBeenPosted(pluginContext *plugin.Context, post *model.Post) {
	_ = p.api.Proxy.Notify(api.NewPostContext(post), api.SubjectPostCreated)
}

func (p *Plugin) ChannelHasBeenCreated(pluginContext *plugin.Context, ch *model.Channel) {
	_ = p.api.Proxy.Notify(api.NewChannelContext(ch), api.SubjectChannelCreated)
}
