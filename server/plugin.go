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

	"github.com/mattermost/mattermost-plugin-apps/apps"
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
	"github.com/mattermost/mattermost-plugin-apps/server/examples/js/aws_hello"
	"github.com/mattermost/mattermost-plugin-apps/server/http"
	"github.com/mattermost/mattermost-plugin-apps/server/http/dialog"
	"github.com/mattermost/mattermost-plugin-apps/server/http/restapi"
)

// const mutexKey = "Cluster_Mutex"

type Plugin struct {
	plugin.MattermostPlugin
	*api.BuildConfig

	mm   *pluginapi.Client
	conf api.Configurator
	aws  aws.Service

	store       *store.Store
	admin       api.Admin
	appservices api.AppServices
	proxy       api.Proxy

	command command.Service
	http    http.Service
}

func NewPlugin(buildConfig *api.BuildConfig) *Plugin {
	return &Plugin{
		BuildConfig: buildConfig,
	}
}

func (p *Plugin) OnActivate() error {
	p.mm = pluginapi.NewClient(p.API)

	botUserID, err := p.mm.Bot.EnsureBot(&model.Bot{
		Username:    api.BotUsername,
		DisplayName: api.BotDisplayName,
		Description: api.BotDescription,
	}, pluginapi.ProfileImagePath("assets/profile.png"))
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot account")
	}

	p.conf = configurator.NewConfigurator(p.mm, p.BuildConfig, botUserID)
	stored := api.StoredConfig{}
	_ = p.mm.Configuration.LoadPluginConfiguration(&stored)
	_ = p.conf.Reconfigure(&stored)
	conf := p.conf.GetConfig()

	p.aws = aws.NewService(&p.mm.Log)
	err = p.aws.Configure(conf)
	if err != nil {
		return errors.Wrap(err, "failed to initialize AWS access")
	}

	p.store = store.New(p.mm, p.conf)
	// manifest store
	mstore := p.store.Manifest()
	mstore.InitBuiltin(
		aws_hello.Manifest(),
		builtin_hello.Manifest(),
	)
	err = mstore.Configure(conf)
	if err != nil {
		return errors.Wrap(err, "failed to initialize the manifest store")
	}
	err = mstore.InitGlobal(p.aws.Client())
	if err != nil {
		return errors.Wrap(err, "failed to initialize the global manifest list from marketplace")
	}
	// app store
	appstore := p.store.App()
	appstore.InitBuiltin(
		builtin_hello.App(),
	)
	err = appstore.Configure(conf)
	if err != nil {
		return errors.Wrap(err, "failed to initialize the app store")
	}

	p.proxy = proxy.NewProxy(p.mm, p.aws, p.conf, p.store)
	p.proxy.AddBuiltinUpstream(builtin_hello.AppID, builtin_hello.New(p.mm))
	p.appservices = appservices.NewAppServices(p.mm, p.conf, p.store)
	p.admin = admin.NewAdmin(p.mm, p.conf, p.store, p.proxy, p.aws, nil)

	p.http = http.NewService(mux.NewRouter(), p.mm, p.conf, p.proxy, p.admin, p.appservices,
		dialog.Init,
		restapi.Init,
		http_hello.Init,
	)
	p.command, err = command.MakeService(p.mm, p.conf, p.proxy, p.admin)
	if err != nil {
		return errors.Wrap(err, "failed to initialize own command handling")
	}

	// if err := p.api.Admin.LoadAppsList(); err != nil {
	// 	mm.Log.Error("Can't load apps list", "err", err.Error())
	// }

	return nil
}

func (p *Plugin) OnConfigurationChange() error {
	if p.conf == nil {
		// pre-activate, nothing to do.
		return nil
	}

	stored := api.StoredConfig{}
	_ = p.mm.Configuration.LoadPluginConfiguration(&stored)

	return p.conf.Reconfigure(&stored,
		p.aws,
		p.store.App(),
		p.store.Manifest())
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	resp, _ := p.command.ExecuteCommand(c, args)
	return resp, nil
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w gohttp.ResponseWriter, req *gohttp.Request) {
	p.http.ServeHTTP(c, w, req)
}

func (p *Plugin) UserHasBeenCreated(pluginContext *plugin.Context, user *model.User) {
	_ = p.proxy.Notify(apps.NewUserContext(user), apps.SubjectUserCreated)
}

func (p *Plugin) UserHasJoinedChannel(pluginContext *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	_ = p.proxy.Notify(apps.NewChannelMemberContext(cm, actingUser), apps.SubjectUserJoinedChannel)
}

func (p *Plugin) UserHasLeftChannel(pluginContext *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	_ = p.proxy.Notify(apps.NewChannelMemberContext(cm, actingUser), apps.SubjectUserLeftChannel)
}

func (p *Plugin) UserHasJoinedTeam(pluginContext *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	_ = p.proxy.Notify(apps.NewTeamMemberContext(tm, actingUser), apps.SubjectUserJoinedTeam)
}

func (p *Plugin) UserHasLeftTeam(pluginContext *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	_ = p.proxy.Notify(apps.NewTeamMemberContext(tm, actingUser), apps.SubjectUserLeftTeam)
}

func (p *Plugin) MessageHasBeenPosted(pluginContext *plugin.Context, post *model.Post) {
	_ = p.proxy.Notify(apps.NewPostContext(post), apps.SubjectPostCreated)
}

func (p *Plugin) ChannelHasBeenCreated(pluginContext *plugin.Context, ch *model.Channel) {
	_ = p.proxy.Notify(apps.NewChannelContext(ch), apps.SubjectChannelCreated)
}
