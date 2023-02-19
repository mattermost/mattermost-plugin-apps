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
	httpIn         *httpin.Service
	httpOut        httpout.Service

	telemetryClient mmtelemetry.Client
}

func NewPlugin(pluginManifest model.Manifest) *Plugin {
	return &Plugin{
		manifest: pluginManifest,
	}
}

func (p *Plugin) OnActivate() (err error) {
	api := config.API{
		Plugin:     p.API,
		Mattermost: pluginapi.NewClient(p.API, p.Driver),
	}
	log := utils.NewPluginLogger(api.Mattermost, nil)

	// Make sure we have the Bot.
	botUserID, err := api.Mattermost.Bot.EnsureBot(&model.Bot{
		Username:    config.BotUsername,
		DisplayName: config.BotDisplayName,
		Description: config.BotDescription,
	}, pluginapi.ProfileImagePath("assets/profile.png"))
	if err != nil {
		return errors.Wrap(err, "failed to ensure bot account")
	}
	log.Debugf("ensured bot @%s: `%s`", config.BotUsername, botUserID)

	// Initialize internalization and telemetrty.
	i18nBundle, err := i18n.InitBundle(p.API, filepath.Join("assets", "i18n"))
	if err != nil {
		return errors.Wrap(err, "failed to load localization files")
	}
	api.I18N = i18nBundle

	p.telemetryClient, err = mmtelemetry.NewRudderClient()
	if err != nil {
		log.WithError(err).Warnw("failed to start telemetry client.")
	}
	api.Telemetry = telemetry.NewTelemetry(nil)

	// Configure the plugin.
	confService, err := config.MakeService(api, p.manifest, botUserID)
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

	// Initialize persistent stores.
	p.store, err = store.MakeService(p.conf, store.SingleWriterCachedStoreKind)
	if err != nil {
		return errors.Wrap(err, "failed to initialize the manifest store")
	}

	if conf.MattermostCloudMode {
		err = p.store.Manifest.InitCloudCatalog(p.conf, p.httpOut)
		if err != nil {
			return errors.Wrap(err, "failed to initialize the manifest store")
		}
	}

	//  Initialize services (API implementations) - session, app services, proxy.
	p.appservices = appservices.NewService(p.store)
	p.sessionService = session.NewService(p.store)
	log.Debugf("initialized API and persistent store")

	// Initialize the app proxy.
	mutex, err := cluster.NewMutex(p.API, store.ClusterMutexKey)
	if err != nil {
		return errors.Wrapf(err, "failed creating cluster mutex")
	}
	p.proxy = proxy.NewService(p.conf, p.store, mutex, p.httpOut, p.sessionService, p.appservices)
	err = p.proxy.Configure(log)
	if err != nil {
		return errors.Wrapf(err, "failed to initialize app proxy")
	}
	log.Debugf("initialized the app proxy")

	// Initialize the built-in "/apps" app.
	p.store.App.InitBuiltin(builtin.App(conf))
	p.proxy.AddBuiltinUpstream(builtin.AppID, builtin.NewBuiltinApp(api, p.proxy, p.appservices, p.httpOut, p.sessionService))
	log.Debugf("initialized the built-in app: use /apps command")

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

	api.Mattermost.Frontend.PublishWebSocketEvent(config.WebSocketEventPluginEnabled, conf.GetPluginVersionInfo(), &model.WebsocketBroadcast{})

	return nil
}

func (p *Plugin) OnDeactivate() error { //nolint:golint,unparam
	conf := p.conf.Get()
	p.conf.API().Mattermost.Frontend.PublishWebSocketEvent(config.WebSocketEventPluginDisabled, conf.GetPluginVersionInfo(), &model.WebsocketBroadcast{})

	if p.telemetryClient != nil {
		err := p.telemetryClient.Close()
		if err != nil {
			p.API.LogWarn("OnDeactivate: failed to close telemetryClient", "error", err.Error())
		}
	}

	return nil
}

func (p *Plugin) OnConfigurationChange() error {
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
	p.conf.API().Telemetry.UpdateTracker(updatedTracker)

	mm := pluginapi.NewClient(p.API, p.Driver)
	var sc config.StoredConfig
	err := mm.Configuration.LoadPluginConfiguration(&sc)
	if err != nil {
		p.API.LogWarn("failed to load updated configuration", "error", err.Error())
		return err
	}

	err = p.conf.Reconfigure(sc, false, p.proxy, p.httpOut)
	if err != nil {
		p.API.LogInfo("failed to reconfigure", "error", err.Error())
		return err
	}
	return nil
}

func (p *Plugin) OnClusterLeaderChanged(isLeader bool) error { //nolint:unparam
	p.conf.OnClusterLeaderChanged(isLeader)
	return nil
}

func (p *Plugin) OnPluginClusterEvent(c *plugin.Context, ev model.PluginClusterEvent) {
	r := p.proxy.NewIncomingRequest("OnPluginClusterEvent", c.RequestId)
	p.store.OnPluginClusterEvent(r, ev)
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w gohttp.ResponseWriter, req *gohttp.Request) {
	// each router.ServeHTTP will create their own incoming.Request using the
	// HTTP headers, otherwise it's difficult to pass r down.
	p.httpIn.ServePluginHTTP(c, w, req)
}

func (p *Plugin) UserHasBeenCreated(c *plugin.Context, user *model.User) {
	r := p.proxy.NewIncomingRequest("UserHasBeenCreated", c.RequestId)
	p.proxy.NotifyUserCreated(r, user.Id)
}

func (p *Plugin) UserHasJoinedChannel(c *plugin.Context, cm *model.ChannelMember, actor *model.User) {
	r := p.proxy.NewIncomingRequest("UserHasJoinedChannel", c.RequestId)
	p.proxy.NotifyUserChannel(r, cm, actor, true)
}

func (p *Plugin) UserHasLeftChannel(c *plugin.Context, cm *model.ChannelMember, actor *model.User) {
	r := p.proxy.NewIncomingRequest("UserHasLeftChannel", c.RequestId)
	p.proxy.NotifyUserChannel(r, cm, actor, false)
}

func (p *Plugin) UserHasJoinedTeam(c *plugin.Context, tm *model.TeamMember, actor *model.User) {
	r := p.proxy.NewIncomingRequest("UserHasJoinedTeam", c.RequestId)
	p.proxy.NotifyUserTeam(r, tm, actor, true)
}

func (p *Plugin) UserHasLeftTeam(c *plugin.Context, tm *model.TeamMember, actor *model.User) {
	r := p.proxy.NewIncomingRequest("UserHasLeftTeam", c.RequestId)
	p.proxy.NotifyUserTeam(r, tm, actor, false)
}

func (p *Plugin) ChannelHasBeenCreated(c *plugin.Context, ch *model.Channel) {
	r := p.proxy.NewIncomingRequest("ChannelHasBeenCreated", c.RequestId)
	p.proxy.NotifyChannelCreated(r, ch.TeamId, ch.Id)
}
