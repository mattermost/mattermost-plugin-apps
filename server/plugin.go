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

	conf config.Service
	log  utils.Logger

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
	mm := pluginapi.NewClient(p.API, p.Driver)

	botUserID, err := mm.Bot.EnsureBot(&model.Bot{
		Username:    config.BotUsername,
		DisplayName: config.BotDisplayName,
		Description: config.BotDescription,
	}, pluginapi.ProfileImagePath("assets/profile.png"))
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot account")
	}

	p.conf = config.NewService(mm, p.BuildConfig, botUserID)
	stored := config.StoredConfig{}
	_ = mm.Configuration.LoadPluginConfiguration(&stored)
	err = p.conf.Reconfigure(stored)
	if err != nil {
		return errors.Wrap(err, "failed to reconfigure configurator on startup")
	}
	conf, _, log := p.conf.Basic()
	p.log = log
	log = log.With("callback", "onactivate")

	mode := "Self-managed"
	if conf.MattermostCloudMode {
		mode = "Mattermost Cloud"
	}
	if conf.DeveloperMode {
		mode += ", Developer Mode"
	}
	log = log.With("mode", mode)

	aws, err := upaws.MakeClient(conf.AWSAccessKey, conf.AWSSecretKey, conf.AWSRegion, p.log)
	if err != nil {
		return errors.Wrap(err, "failed to initialize AWS access")
	}
	log = log.With(
		"aws_region", conf.AWSRegion,
		"aws_bucket", conf.AWSS3Bucket,
		"aws_access", utils.LastN(conf.AWSAccessKey, 7),
		"aws_secret", utils.LastN(conf.AWSSecretKey, 4))

	p.httpOut = httpout.NewService(p.conf)

	p.store = store.NewService(p.conf, aws, conf.AWSS3Bucket)
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

	mutex, err := cluster.NewMutex(p.API, config.KVClusterMutexKey)
	if err != nil {
		return errors.Wrapf(err, "failed creating cluster mutex")
	}
	p.proxy = proxy.NewService(p.conf, aws, conf.AWSS3Bucket, p.store, mutex, p.httpOut)

	p.appservices = appservices.NewService(p.conf, p.store)

	p.httpIn = httpin.NewService(mux.NewRouter(), p.conf, p.proxy, p.appservices,
		dialog.Init,
		restapi.Init,
		gateway.Init,
		http_hello.Init,
	)

	p.command, err = command.MakeService(p.conf, p.proxy, p.httpOut)
	if err != nil {
		return errors.Wrap(err, "failed to initialize own command handling")
	}

	if conf.MattermostCloudMode {
		err = p.proxy.SynchronizeInstalledApps()
		if err != nil {
			log.WithError(err).Errorf("Failed to synchronize apps metadata")
		} else {
			log.Debugf("Synchronized the installed apps metadata")
		}
	}
	log.Infof("Plugin activated")

	return nil
}

func (p *Plugin) OnConfigurationChange() error {
	if p.conf == nil {
		// pre-activate, nothing to do.
		return nil
	}

	mm := pluginapi.NewClient(p.API, p.Driver)
	stored := config.StoredConfig{}
	_ = mm.Configuration.LoadPluginConfiguration(&stored)

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
	cc := p.conf.Get().SetContextDefaults(&apps.Context{
		UserID: user.Id,
		ExpandedContext: apps.ExpandedContext{
			User: user,
		},
		Locale: utils.GetLocale(p.conf.MattermostAPI(), p.conf.MattermostConfig().Config(), user.Id),
	})
	err := p.proxy.Notify(cc, apps.SubjectUserCreated)
	if err != nil {
		p.log.WithError(err).Debugf("Error handling UserHasBeenCreated")
	}
}

func (p *Plugin) UserHasJoinedChannel(pluginContext *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	err := p.proxy.NotifyUserHasJoinedChannel(p.newChannelMemberContext(cm, actingUser))
	if err != nil {
		p.log.WithError(err).Debugf("Error handling UserHasJoinedChannel")
	}
}

func (p *Plugin) UserHasLeftChannel(pluginContext *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	err := p.proxy.NotifyUserHasLeftChannel(p.newChannelMemberContext(cm, actingUser))
	if err != nil {
		p.log.WithError(err).Debugf("Error handling UserHasLeftChannel")
	}
}

func (p *Plugin) UserHasJoinedTeam(pluginContext *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	err := p.proxy.NotifyUserHasJoinedTeam(p.newTeamMemberContext(tm, actingUser))
	if err != nil {
		p.log.WithError(err).Debugf("Error handling UserHasJoinedTeam")
	}
}

func (p *Plugin) UserHasLeftTeam(pluginContext *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	err := p.proxy.NotifyUserHasLeftTeam(p.newTeamMemberContext(tm, actingUser))
	if err != nil {
		p.log.WithError(err).Debugf("Error handling UserHasLeftTeam")
	}
}

func (p *Plugin) MessageHasBeenPosted(pluginContext *plugin.Context, post *model.Post) {
	shouldProcessMessage, err := p.Helpers.ShouldProcessMessage(post, plugin.BotID(p.conf.Get().BotUserID))
	if err != nil {
		p.log.WithError(err).Errorf("Error while checking if the message should be processed")
		return
	}

	if !shouldProcessMessage {
		return
	}

	err = p.proxy.NotifyMessageHasBeenPosted(post, p.newPostCreatedContext(post))
	if err != nil {
		p.log.WithError(err).Debugf("Error handling MessageHasBeenPosted")
	}
}

func (p *Plugin) ChannelHasBeenCreated(pluginContext *plugin.Context, ch *model.Channel) {
	cc := p.conf.Get().SetContextDefaults(&apps.Context{
		UserAgentContext: apps.UserAgentContext{
			TeamID:    ch.TeamId,
			ChannelID: ch.Id,
		},
		UserID: ch.CreatorId,
		Locale: utils.GetLocale(p.conf.MattermostAPI(), p.conf.MattermostConfig().Config(), ch.CreatorId),
		ExpandedContext: apps.ExpandedContext{
			Channel: ch,
		},
	})
	_ = p.proxy.Notify(cc, apps.SubjectChannelCreated)
}

func (p *Plugin) newPostCreatedContext(post *model.Post) *apps.Context {
	return p.conf.Get().SetContextDefaults(&apps.Context{
		UserAgentContext: apps.UserAgentContext{
			PostID:     post.Id,
			RootPostID: post.RootId,
			ChannelID:  post.ChannelId,
		},
		UserID: post.UserId,
		Locale: utils.GetLocale(p.conf.MattermostAPI(), p.conf.MattermostConfig().Config(), post.UserId),
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
	return p.conf.Get().SetContextDefaults(&apps.Context{
		UserAgentContext: apps.UserAgentContext{
			TeamID: tm.TeamId,
		},
		ActingUserID: actingUserID,
		UserID:       tm.UserId,
		Locale:       utils.GetLocale(p.conf.MattermostAPI(), p.conf.MattermostConfig().Config(), actingUserID),
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
	return p.conf.Get().SetContextDefaults(&apps.Context{
		UserAgentContext: apps.UserAgentContext{
			ChannelID: cm.ChannelId,
		},
		ActingUserID: actingUserID,
		UserID:       cm.UserId,
		Locale:       utils.GetLocale(p.conf.MattermostAPI(), p.conf.MattermostConfig().Config(), actingUserID),
		ExpandedContext: apps.ExpandedContext{
			ActingUser: actingUser,
		},
	})
}
