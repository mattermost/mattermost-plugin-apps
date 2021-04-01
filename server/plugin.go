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
	"github.com/mattermost/mattermost-plugin-apps/aws"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/command"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/examples/go/hello/http_hello"
	"github.com/mattermost/mattermost-plugin-apps/server/http"
	"github.com/mattermost/mattermost-plugin-apps/server/http/dialog"
	"github.com/mattermost/mattermost-plugin-apps/server/http/restapi"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

type Plugin struct {
	plugin.MattermostPlugin
	config.BuildConfig

	mm   *pluginapi.Client
	conf config.Service
	aws  aws.Client

	store       *store.Service
	appservices appservices.Service
	proxy       proxy.Service

	command command.Service
	http    http.Service
}

func NewPlugin(buildConfig config.BuildConfig) *Plugin {
	return &Plugin{
		BuildConfig: buildConfig,
	}
}

func (p *Plugin) OnActivate() error {
	p.mm = pluginapi.NewClient(p.API)

	botUserID, err := p.mm.Bot.EnsureBot(&model.Bot{
		Username:    config.BotUsername,
		DisplayName: config.BotDisplayName,
		Description: config.BotDescription,
	}, pluginapi.ProfileImagePath("assets/profile.png"))
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot account")
	}

	p.conf = config.NewService(p.mm, p.BuildConfig, botUserID)
	stored := config.StoredConfig{}
	_ = p.mm.Configuration.LoadPluginConfiguration(&stored)
	err = p.conf.Reconfigure(stored)
	if err != nil {
		return errors.Wrap(err, "failed to reconfigure configurator on startup")
	}

	accessKey := os.Getenv("APPS_INVOKE_AWS_ACCESS_KEY")
	if accessKey == "" {
		p.mm.Log.Warn("APPS_INVOKE_AWS_ACCESS_KEY is not set. AWS apps won't work.")
	}
	secretKey := os.Getenv("APPS_INVOKE_AWS_SECRET_KEY")
	if secretKey == "" {
		p.mm.Log.Warn("APPS_INVOKE_AWS_SECRET_KEY is not set. AWS apps won't work.")
	}
	p.aws, err = aws.MakeClient(accessKey, secretKey, &p.mm.Log)
	if err != nil {
		return errors.Wrap(err, "failed to initialize AWS access")
	}

	p.store = store.NewService(p.mm, p.conf)
	// manifest store
	conf := p.conf.GetConfig()
	mstore := p.store.Manifest
	mstore.Configure(conf)
	// TODO: uses the default bucket name, do we need it customizeable?
	manifestBucket := apps.S3BucketNameWithDefaults("")
	err = mstore.InitGlobal(p.aws, manifestBucket)
	if err != nil {
		return errors.Wrap(err, "failed to initialize the global manifest list from marketplace")
	}
	// app store
	appstore := p.store.App
	appstore.Configure(conf)

	// TODO: uses the default bucket name, same as for the manifests do we need
	// it customizeable?
	assetBucket := apps.S3BucketNameWithDefaults("")
	mutex, err := cluster.NewMutex(p.API, config.KeyClusterMutex)
	if err != nil {
		return errors.Wrapf(err, "failed creating cluster mutex")
	}

	p.proxy = proxy.NewService(p.mm, p.aws, p.conf, p.store, assetBucket, mutex)

	p.appservices = appservices.NewService(p.mm, p.conf, p.store)

	p.http = http.NewService(mux.NewRouter(), p.mm, p.conf, p.proxy, p.appservices,
		dialog.Init,
		restapi.Init,
		http_hello.Init,
	)
	p.command, err = command.MakeService(p.mm, p.conf, p.proxy)
	if err != nil {
		return errors.Wrap(err, "failed to initialize own command handling")
	}

	err = p.proxy.SynchronizeInstalledApps()
	if err != nil {
		p.mm.Log.Error("failed to update apps", "err", err.Error())
	}

	return nil
}

func (p *Plugin) OnConfigurationChange() error {
	if p.conf == nil {
		// pre-activate, nothing to do.
		return nil
	}

	stored := config.StoredConfig{}
	_ = p.mm.Configuration.LoadPluginConfiguration(&stored)

	return p.conf.Reconfigure(stored, p.store.App, p.store.Manifest)
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
