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
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/builtin"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin/gateway"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin/restapi"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Plugin struct {
	plugin.MattermostPlugin
	config.BuildConfig

	mm   *pluginapi.Client
	conf config.Service
	log  utils.Logger

	store       *store.Service
	appservices appservices.Service
	proxy       proxy.Service

	httpIn  httpin.Service
	httpOut httpout.Service
}

func NewPlugin(buildConfig config.BuildConfig) *Plugin {
	return &Plugin{
		BuildConfig: buildConfig,
	}
}

func (p *Plugin) OnActivate() (err error) {
	p.mm = pluginapi.NewClient(p.API, p.Driver)
	p.log = utils.NewPluginLogger(p.mm)

	defer func() {
		if err != nil {
			p.log.WithError(err).Errorf("Failed to activate")
		}
	}()

	botUserID, err := p.mm.Bot.EnsureBot(&model.Bot{
		Username:    config.BotUsername,
		DisplayName: config.BotDisplayName,
		Description: config.BotDescription,
	}, pluginapi.ProfileImagePath("assets/profile.png"))
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot account")
	}

	p.conf = config.NewService(p.mm, p.log, p.BuildConfig, botUserID)
	stored := config.StoredConfig{}
	_ = p.mm.Configuration.LoadPluginConfiguration(&stored)
	err = p.conf.Reconfigure(stored)
	if err != nil {
		return errors.Wrap(err, "failed to load initial configuration")
	}
	conf := p.conf.GetConfig()
	mode := "Self-managed"
	if conf.MattermostCloudMode {
		mode = "Mattermost Cloud"
	}
	if conf.DeveloperMode {
		mode += ", Developer Mode"
	}
	p.log.Debugf("Initialized config service: %s", mode)

	p.httpOut = httpout.NewService(p.conf)
	p.log.Debugf("Initialized outgoing HTTP")

	p.store, err = store.MakeService(p.mm, p.log, p.conf, p.httpOut)
	if err != nil {
		return errors.Wrap(err, "failed to initialize persistent store")
	}
	p.store.App.InitBuiltin(builtin.App(conf))
	p.log.Debugf("Initialized persistent store")

	mutex, err := cluster.NewMutex(p.API, config.KVClusterMutexKey)
	if err != nil {
		return errors.Wrapf(err, "failed creating cluster mutex")
	}
	p.proxy = proxy.NewService(p.mm, p.log, p.conf, p.store, mutex, p.httpOut)
	err = p.proxy.Configure(conf)
	if err != nil {
		return errors.Wrapf(err, "failed to initialize app proxy service")
	}
	p.proxy.AddBuiltinUpstream(
		builtin.AppID,
		builtin.NewBuiltinApp(p.mm, p.log, p.conf, p.proxy, p.store, p.httpOut),
	)
	p.log.Debugf("Initialized the app proxy")

	p.appservices = appservices.NewService(p.mm, p.conf, p.store)
	p.log.Debugf("Initialized the app REST APIs")

	p.httpIn = httpin.NewService(mux.NewRouter(), p.mm, p.log, p.conf, p.proxy, p.appservices,
		restapi.Init,
		gateway.Init,
	)
	p.log.Debugf("Initialized incoming HTTP")

	if conf.MattermostCloudMode {
		err = p.proxy.SynchronizeInstalledApps()
		if err != nil {
			p.log.WithError(err).Errorf("Failed to synchronize apps metadata")
		} else {
			p.log.Debugf("Synchronized the installed apps metadata")
		}
	}

	return nil
}

func (p *Plugin) OnConfigurationChange() (err error) {
	defer func() {
		if err != nil {
			p.log.WithError(err).Errorf("Failed to reconfigure")
		}
	}()

	if p.conf == nil {
		// pre-activate, nothing to do.
		return nil
	}

	stored := config.StoredConfig{}
	_ = p.mm.Configuration.LoadPluginConfiguration(&stored)

	return p.conf.Reconfigure(stored, p.store.App, p.store.Manifest, p.proxy)
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w gohttp.ResponseWriter, req *gohttp.Request) {
	p.httpIn.ServeHTTP(c, w, req)
}

func (p *Plugin) UserHasBeenCreated(pluginContext *plugin.Context, user *model.User) {
	cc := apps.Context{
		UserID: user.Id,
		ExpandedContext: apps.ExpandedContext{
			User: user,
		},
	}
	_ = p.proxy.Notify(cc, apps.SubjectUserCreated)
}

func (p *Plugin) UserHasJoinedChannel(pluginContext *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	_ = p.proxy.Notify(p.newChannelMemberContext(cm), apps.SubjectUserJoinedChannel)
}

func (p *Plugin) UserHasLeftChannel(pluginContext *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	_ = p.proxy.Notify(p.newChannelMemberContext(cm), apps.SubjectUserLeftChannel)
}

func (p *Plugin) UserHasJoinedTeam(pluginContext *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	_ = p.proxy.Notify(p.newTeamMemberContext(tm), apps.SubjectUserJoinedTeam)
}

func (p *Plugin) UserHasLeftTeam(pluginContext *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	_ = p.proxy.Notify(p.newTeamMemberContext(tm), apps.SubjectUserLeftTeam)
}

func (p *Plugin) MessageHasBeenPosted(pluginContext *plugin.Context, post *model.Post) {
	shouldProcessMessage, err := p.Helpers.ShouldProcessMessage(post, plugin.BotID(p.conf.GetConfig().BotUserID))
	if err != nil {
		p.log.WithError(err).Errorf("Error while checking if the message should be processed")
		return
	}

	if !shouldProcessMessage {
		return
	}

	cc := apps.Context{
		UserAgentContext: apps.UserAgentContext{
			PostID:     post.Id,
			RootPostID: post.RootId,
			ChannelID:  post.ChannelId,
		},
		UserID: post.UserId,
		ExpandedContext: apps.ExpandedContext{
			Post: post,
		},
	}
	_ = p.proxy.Notify(cc, apps.SubjectPostCreated)
}

func (p *Plugin) ChannelHasBeenCreated(pluginContext *plugin.Context, ch *model.Channel) {
	cc := apps.Context{
		UserAgentContext: apps.UserAgentContext{
			TeamID:    ch.TeamId,
			ChannelID: ch.Id,
		},
		UserID: ch.CreatorId,
		ExpandedContext: apps.ExpandedContext{
			Channel: ch,
		},
	}
	_ = p.proxy.Notify(cc, apps.SubjectChannelCreated)
}

func (p *Plugin) newTeamMemberContext(tm *model.TeamMember) apps.Context {
	return apps.Context{
		UserAgentContext: apps.UserAgentContext{
			TeamID: tm.TeamId,
		},
		UserID: tm.UserId,
	}
}

func (p *Plugin) newChannelMemberContext(cm *model.ChannelMember) apps.Context {
	return apps.Context{
		UserAgentContext: apps.UserAgentContext{
			ChannelID: cm.ChannelId,
		},
		UserID: cm.UserId,
	}
}
