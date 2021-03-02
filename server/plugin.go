// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	gohttp "net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/cluster"
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
	"github.com/mattermost/mattermost-plugin-apps/server/http"
	"github.com/mattermost/mattermost-plugin-apps/server/http/dialog"
	"github.com/mattermost/mattermost-plugin-apps/server/http/restapi"
)

const mutexKey = "Cluster_Mutex"

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

	accessKey := os.Getenv("APPS_INVOKE_AWS_ACCESS_KEY")
	if accessKey == "" {
		mm.Log.Warn("APPS_INVOKE_AWS_ACCESS_KEY is not set. AWS apps won't work.")
	}
	secretKey := os.Getenv("APPS_INVOKE_AWS_SECRET_KEY")
	if secretKey == "" {
		mm.Log.Warn("APPS_INVOKE_AWS_SECRET_KEY is not set. AWS apps won't work.")
	}

	awsClient := aws.NewAWSClient(accessKey, secretKey, &mm.Log)

	conf := configurator.NewConfigurator(mm, awsClient, p.BuildConfig, botUserID)
	_ = conf.RefreshConfig(&stored)
	store := store.New(mm, conf)
	proxy := proxy.NewProxy(mm, awsClient, conf, store)

	mutex, err := cluster.NewMutex(p.API, mutexKey)
	if err != nil {
		return errors.Wrapf(err, "failed creating cluster mutex")
	}

	p.api = &api.Service{
		Mattermost:   mm,
		Configurator: conf,
		Proxy:        proxy,
		AppServices:  appservices.NewAppServices(mm, conf, store),
		Admin:        admin.NewAdmin(mm, conf, store, proxy, awsClient, mutex),
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

	if err := p.api.Admin.LoadAppsList(); err != nil {
		mm.Log.Error("Can't load apps list", "err", err.Error())
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
	_ = p.api.Proxy.Notify(apps.NewUserContext(user), apps.SubjectUserCreated)
}

func (p *Plugin) UserHasJoinedChannel(pluginContext *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	_ = p.api.Proxy.Notify(apps.NewChannelMemberContext(cm, actingUser), apps.SubjectUserJoinedChannel)
}

func (p *Plugin) UserHasLeftChannel(pluginContext *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	_ = p.api.Proxy.Notify(apps.NewChannelMemberContext(cm, actingUser), apps.SubjectUserLeftChannel)
}

func (p *Plugin) UserHasJoinedTeam(pluginContext *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	_ = p.api.Proxy.Notify(apps.NewTeamMemberContext(tm, actingUser), apps.SubjectUserJoinedTeam)
}

func (p *Plugin) UserHasLeftTeam(pluginContext *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	_ = p.api.Proxy.Notify(apps.NewTeamMemberContext(tm, actingUser), apps.SubjectUserLeftTeam)
}

func (p *Plugin) MessageHasBeenPosted(pluginContext *plugin.Context, post *model.Post) {
	_ = p.api.Proxy.Notify(apps.NewPostContext(post), apps.SubjectPostCreated)
}

func (p *Plugin) ChannelHasBeenCreated(pluginContext *plugin.Context, ch *model.Channel) {
	_ = p.api.Proxy.Notify(apps.NewChannelContext(ch), apps.SubjectChannelCreated)
}
