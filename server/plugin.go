// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	gohttp "net/http"
	"path/filepath"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/cluster"
	mmtelemetry "github.com/mattermost/mattermost-plugin-api/experimental/telemetry"
	"github.com/mattermost/mattermost-plugin-api/i18n"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/builtin"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/session"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/telemetry"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Plugin struct {
	plugin.MattermostPlugin
	manifest model.Manifest

	conf config.Service

	store          *store.Service
	appservices    appservices.Service
	proxy          proxy.Service
	sessionService session.Service

	httpIn  *httpin.Service
	httpOut httpout.Service

	telemetryClient mmtelemetry.Client
	tracker         *telemetry.Telemetry
}

func NewPlugin(pluginManifest model.Manifest) *Plugin {
	return &Plugin{
		manifest: pluginManifest,
	}
}

func (p *Plugin) OnActivate() (err error) {
	mm := pluginapi.NewClient(p.API, p.Driver)
	log := utils.NewPluginLogger(mm, nil)

	// Make sure we have the Bot.
	botUserID, err := mm.Bot.EnsureBot(&model.Bot{
		Username:    config.BotUsername,
		DisplayName: config.BotDisplayName,
		Description: config.BotDescription,
	}, pluginapi.ProfileImagePath("assets/profile.png"))
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot account")
	}
	log.Debugw("ensured bot", "id", botUserID, "username", config.BotUsername)

	// Initialize internalization and telemetrty.
	i18nBundle, err := i18n.InitBundle(p.API, filepath.Join("assets", "i18n"))
	if err != nil {
		return errors.Wrap(err, "failed to load localization files")
	}

	p.telemetryClient, err = mmtelemetry.NewRudderClient()
	if err != nil {
		log.WithError(err).Warnw("failed to start telemetry client.")
	}
	p.tracker = telemetry.NewTelemetry(nil)

	// Configure the plugin.
	confService, err := config.NewService(mm, p.manifest, botUserID, p.tracker, i18nBundle, log)
	if err != nil {
		log.WithError(err).Infow("failed to load initial configuration")
		return errors.Wrap(err, "failed to load initial configuration")
	}
	p.conf = confService
	log = p.conf.NewBaseLogger()

	conf := p.conf.Get()
	log.With(conf).Debugw("configured the plugin.")

	// Initialize outgoing HTTP.
	p.httpOut = httpout.NewService(p.conf)

	// Initialize persistent storage. Also initialize the app API and the
	// session services, both need the persisitent store.
	p.store, err = store.MakeService(p.conf, p.httpOut)
	if err != nil {
		return errors.Wrap(err, "failed to initialize persistent store")
	}
	p.store.App.InitBuiltin(builtin.App(conf))
	p.appservices = appservices.NewService(p.store)
	p.sessionService = session.NewService(mm, p.store)
	log.Debugf("initialized API and persistent store")

	// Initialize the app proxy.
	mutex, err := cluster.NewMutex(p.API, store.KVClusterMutexKey)
	if err != nil {
		return errors.Wrapf(err, "failed creating cluster mutex")
	}
	p.proxy = proxy.NewService(p.conf, p.store, mutex, p.httpOut, p.sessionService, p.appservices)
	err = p.proxy.Configure(conf, log)
	if err != nil {
		return errors.Wrapf(err, "failed to initialize app proxy")
	}
	p.proxy.AddBuiltinUpstream(
		builtin.AppID,
		builtin.NewBuiltinApp(p.conf, p.proxy, p.appservices, p.httpOut, p.sessionService),
	)
	log.Debugf("initialized the app proxy")

	p.httpIn = httpin.NewService(p.proxy, p.appservices, p.conf)
	log.Debugf("initialized incoming HTTP")

	if conf.MattermostCloudMode {
		err = p.proxy.SynchronizeInstalledApps()
		if err != nil {
			log.WithError(err).Errorf("failed to synchronize apps metadata")
		} else {
			log.Debugf("Synchronized the installed apps metadata")
		}
	}
	log.Infof("activated")

	p.conf.MattermostAPI().Frontend.PublishWebSocketEvent(config.WebSocketEventPluginEnabled, conf.GetPluginVersionInfo(), &model.WebsocketBroadcast{})

	return nil
}

func (p *Plugin) OnDeactivate() error { //nolint:golint,unparam
	conf := p.conf.Get()
	p.conf.MattermostAPI().Frontend.PublishWebSocketEvent(config.WebSocketEventPluginDisabled, conf.GetPluginVersionInfo(), &model.WebsocketBroadcast{})

	if p.telemetryClient != nil {
		err := p.telemetryClient.Close()
		if err != nil {
			p.API.LogWarn("OnDeactivate: failed to close telemetryClient", "error", err.Error())
		}
	}

	return nil
}

func (p *Plugin) OnConfigurationChange() (err error) {
	if p.conf == nil {
		// pre-activate, nothing to do.
		return nil
	}

	enableDiagnostics := false
	if config := p.API.GetConfig(); config != nil {
		if configValue := config.LogSettings.EnableDiagnostics; configValue != nil {
			enableDiagnostics = *configValue
		}
	}
	updatedTracker := mmtelemetry.NewTracker(p.telemetryClient, p.API.GetDiagnosticId(), p.API.GetServerVersion(), manifest.Id, manifest.Version, "appsFramework", enableDiagnostics)
	p.tracker.UpdateTracker(updatedTracker)

	mm := pluginapi.NewClient(p.API, p.Driver)
	cm := mm.Configuration.GetPluginConfig()

	err = p.conf.Reconfigure(cm, nil, p.store.App, p.store.Manifest, p.proxy)
	if err != nil {
		p.API.LogInfo("failed to reconfigure", "error", err.Error())
		return err
	}
	return nil
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w gohttp.ResponseWriter, req *gohttp.Request) {
	p.httpIn.ServePluginHTTP(c, w, req)
}

func (p *Plugin) UserHasBeenCreated(_ *plugin.Context, user *model.User) {
	p.proxy.NotifyUserCreated(user.Id)
}

func (p *Plugin) UserHasJoinedChannel(_ *plugin.Context, cm *model.ChannelMember, actor *model.User) {
	p.proxy.NotifyUserChannel(cm, actor, true)
}

func (p *Plugin) UserHasLeftChannel(_ *plugin.Context, cm *model.ChannelMember, actor *model.User) {
	p.proxy.NotifyUserChannel(cm, actor, false)
}

func (p *Plugin) UserHasJoinedTeam(_ *plugin.Context, tm *model.TeamMember, actor *model.User) {
	p.proxy.NotifyUserTeam(tm, actor, true)
}

func (p *Plugin) UserHasLeftTeam(_ *plugin.Context, tm *model.TeamMember, actor *model.User) {
	p.proxy.NotifyUserTeam(tm, actor, false)
}

func (p *Plugin) ChannelHasBeenCreated(_ *plugin.Context, ch *model.Channel) {
	p.proxy.NotifyChannelCreated(ch.TeamId, ch.Id)
}
