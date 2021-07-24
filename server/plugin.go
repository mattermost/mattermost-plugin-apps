// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	gohttp "net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/cluster"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/examples/go/hello/http_hello"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/command"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin/dialog"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin/gateway"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin/restapi"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Plugin struct {
	plugin.MattermostPlugin
	config.BuildConfig

	mm   *pluginapi.Client
	conf config.Service
	aws  upaws.Client

	store       *store.Service
	appservices appservices.Service
	proxy       proxy.Service

	command command.Service

	httpIn  httpin.Service
	httpOut httpout.Service
}

func NewPlugin(buildConfig config.BuildConfig) *Plugin {
	return &Plugin{
		BuildConfig: buildConfig,
	}
}

func (p *Plugin) OnActivate() error {
	p.mm = pluginapi.NewClient(p.API, p.Driver)

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
	conf := p.conf.GetConfig()
	mode := "Self-managed"
	if conf.MattermostCloudMode {
		mode = "Mattermost Cloud"
	}
	if conf.DeveloperMode {
		mode += ", Developer Mode"
	}
	p.mm.Log.Debug("Initialized config service: " + mode)

	p.aws, err = upaws.MakeClient(conf.AWSAccessKey, conf.AWSSecretKey, conf.AWSRegion, &p.mm.Log)
	if err != nil {
		return errors.Wrap(err, "failed to initialize AWS access")
	}
	p.mm.Log.Debug("Initialized AWS Client",
		"region", conf.AWSRegion, "bucket", conf.AWSS3Bucket,
		"access", utils.LastN(conf.AWSAccessKey, 7), "secret", utils.LastN(conf.AWSSecretKey, 4))

	p.httpOut = httpout.NewService(p.conf)
	p.mm.Log.Debug("Initialized outgoing HTTP")

	p.store = store.NewService(p.mm, p.conf, p.aws, conf.AWSS3Bucket)
	// manifest store
	mstore := p.store.Manifest
	mstore.Configure(conf)
	if conf.MattermostCloudMode {
		err = mstore.InitGlobal(p.httpOut)
		if err != nil {
			return errors.Wrap(err, "failed to initialize the global manifest list from marketplace")
		}
	}
	// app store
	appstore := p.store.App
	appstore.Configure(conf)
	p.mm.Log.Debug("Initialized the persistent store")

	mutex, err := cluster.NewMutex(p.API, config.KVClusterMutexKey)
	if err != nil {
		return errors.Wrapf(err, "failed creating cluster mutex")
	}
	p.proxy = proxy.NewService(p.mm, p.aws, p.conf, p.store, conf.AWSS3Bucket, mutex, p.httpOut)
	p.mm.Log.Debug("Initialized the app proxy")

	p.appservices = appservices.NewService(p.mm, p.conf, p.store)
	p.mm.Log.Debug("Initialized the app REST APIs")

	p.httpIn = httpin.NewService(mux.NewRouter(), p.mm, p.conf, p.proxy, p.appservices,
		dialog.Init,
		restapi.Init,
		gateway.Init,
		http_hello.Init,
	)
	p.mm.Log.Debug("Initialized incoming HTTP")

	p.command, err = command.MakeService(p.mm, p.conf, p.proxy, p.httpOut)
	if err != nil {
		return errors.Wrap(err, "failed to initialize own command handling")
	}
	p.mm.Log.Debug("Initialized slash commands")

	if conf.MattermostCloudMode {
		err = p.proxy.SynchronizeInstalledApps()
		if err != nil {
			p.mm.Log.Error("Failed to synchronize apps metadata", "err", err.Error())
		} else {
			p.mm.Log.Debug("Synchronized the installed apps metadata")
		}
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

	return p.conf.Reconfigure(stored, p.store.App, p.store.Manifest, p.command)
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	resp, _ := p.command.ExecuteCommand(c, args)
	return resp, nil
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w gohttp.ResponseWriter, req *gohttp.Request) {
	p.httpIn.ServeHTTP(c, w, req)
}

func (p *Plugin) UserHasBeenCreated(pluginContext *plugin.Context, user *model.User) {
	cc := p.conf.GetConfig().SetContextDefaults(&apps.Context{
		UserID: user.Id,
		ExpandedContext: apps.ExpandedContext{
			User: user,
		},
	})
	_ = p.proxy.Notify(cc, apps.SubjectUserCreated)
}

func (p *Plugin) UserHasJoinedChannel(pluginContext *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	_ = p.proxy.NotifyUserHasJoinedChannel(p.newChannelMemberContext(cm, actingUser))
}

func (p *Plugin) UserHasLeftChannel(pluginContext *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	_ = p.proxy.NotifyUserHasLeftChannel(p.newChannelMemberContext(cm, actingUser))
}

func (p *Plugin) UserHasJoinedTeam(pluginContext *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	_ = p.proxy.Notify(p.newTeamMemberContext(tm, actingUser), apps.SubjectUserJoinedTeam)
}

func (p *Plugin) UserHasLeftTeam(pluginContext *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	_ = p.proxy.Notify(p.newTeamMemberContext(tm, actingUser), apps.SubjectUserLeftTeam)
}

func (p *Plugin) MessageHasBeenPosted(pluginContext *plugin.Context, post *model.Post) {
	shouldProcessMessage, err := p.Helpers.ShouldProcessMessage(post, plugin.BotID(p.conf.GetConfig().BotUserID))
	if err != nil {
		p.mm.Log.Error("Error while checking if the message should be processed", "err", err.Error())
		return
	}

	if !shouldProcessMessage {
		return
	}

	p.proxy.NotifyMessageHasBeenPosted(post, p.newPostCreatedContext(post))
}

func (p *Plugin) ChannelHasBeenCreated(pluginContext *plugin.Context, ch *model.Channel) {
	cc := p.conf.GetConfig().SetContextDefaults(&apps.Context{
		UserAgentContext: apps.UserAgentContext{
			TeamID:    ch.TeamId,
			ChannelID: ch.Id,
		},
		UserID: ch.CreatorId,
		ExpandedContext: apps.ExpandedContext{
			Channel: ch,
		},
	})
	_ = p.proxy.Notify(cc, apps.SubjectChannelCreated)
}

func (p *Plugin) newPostCreatedContext(post *model.Post) *apps.Context {
	return p.conf.GetConfig().SetContextDefaults(&apps.Context{
		UserAgentContext: apps.UserAgentContext{
			PostID:     post.Id,
			RootPostID: post.RootId,
			ChannelID:  post.ChannelId,
		},
		UserID: post.UserId,
		ExpandedContext: apps.ExpandedContext{
			Post: post,
		},
	})
}

func (p *Plugin) newTeamMemberContext(tm *model.TeamMember, actingUser *model.User) *apps.Context {
	actingUserID := ""
	if actingUser != nil {
		actingUserID = actingUser.Id
	}
	return p.conf.GetConfig().SetContextDefaults(&apps.Context{
		UserAgentContext: apps.UserAgentContext{
			TeamID: tm.TeamId,
		},
		ActingUserID: actingUserID,
		UserID:       tm.UserId,
		ExpandedContext: apps.ExpandedContext{
			ActingUser: actingUser,
		},
	})
}

func (p *Plugin) newChannelMemberContext(cm *model.ChannelMember, actingUser *model.User) *apps.Context {
	actingUserID := ""
	if actingUser != nil {
		actingUserID = actingUser.Id
	}
	return p.conf.GetConfig().SetContextDefaults(&apps.Context{
		UserAgentContext: apps.UserAgentContext{
			ChannelID: cm.ChannelId,
		},
		ActingUserID: actingUserID,
		UserID:       cm.UserId,
		ExpandedContext: apps.ExpandedContext{
			ActingUser: actingUser,
		},
	})
}
