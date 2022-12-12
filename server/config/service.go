package config

import (
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/i18n"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/services/configservice"

	"github.com/mattermost/mattermost-plugin-apps/server/telemetry"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Configurable interface {
	Configure(Config, utils.Logger) error
}

type Service interface {
	Get() Config
	MattermostAPI() *pluginapi.Client
	MattermostConfig() configservice.ConfigService
	I18N() *i18n.Bundle
	Telemetry() *telemetry.Telemetry
	NewBaseLogger() utils.Logger

	Reconfigure(StoredConfig, utils.Logger, ...Configurable) error
	StoreConfig(StoredConfig, utils.Logger) error
}

var _ Service = (*service)(nil)

type service struct {
	pluginManifest model.Manifest
	botUserID      string
	mm             *pluginapi.Client
	i18n           *i18n.Bundle
	telemetry      *telemetry.Telemetry

	lock             *sync.RWMutex
	conf             *Config
	mattermostConfig *model.Config
}

func NewService(mm *pluginapi.Client, pliginManifest model.Manifest, botUserID string, telemetry *telemetry.Telemetry, i18nBundle *i18n.Bundle) Service {
	return &service{
		pluginManifest: pliginManifest,
		botUserID:      botUserID,
		mm:             mm,
		lock:           &sync.RWMutex{},
		i18n:           i18nBundle,
		telemetry:      telemetry,
	}
}

func (s *service) Get() Config {
	s.lock.RLock()
	conf := s.conf
	s.lock.RUnlock()

	if conf == nil {
		return s.newConfig()
	}
	return *conf
}

func (s *service) MattermostAPI() *pluginapi.Client {
	return s.mm
}

func (s *service) I18N() *i18n.Bundle {
	return s.i18n
}

func (s *service) Telemetry() *telemetry.Telemetry {
	return s.telemetry
}

func (s *service) MattermostConfig() configservice.ConfigService {
	s.lock.RLock()
	mmconf := s.mattermostConfig
	s.lock.RUnlock()

	if mmconf == nil {
		mmconf = s.reloadMattermostConfig()
	}
	return &mattermostConfigService{
		mmconf: mmconf,
	}
}

func (s *service) reloadMattermostConfig() *model.Config {
	mmconf := s.mm.Configuration.GetConfig()

	s.lock.Lock()
	s.mattermostConfig = mmconf
	s.lock.Unlock()

	return mmconf
}

func (s *service) getMattermostLicense(log utils.Logger) *model.License {
	// GetLicense silently drops an RPC error
	// (https://github.com/mattermost/mattermost-server/blob/fc75b72bbabf7fabfad24b9e1e4c321ca9b9b7f1/plugin/client_rpc_generated.go#L864).
	// When running in Mattermost cloud we must not fall back to the on-prem mode, so in case we get a nil retry once.
	license := s.mm.System.GetLicense()
	if license == nil {
		license = s.mm.System.GetLicense()
		if license == nil {
			log.Debugf("failed to fetch license twice. May incorrectly default to on-prem mode")
		}
	}

	return license
}

func (s *service) Reconfigure(newStoredConfig StoredConfig, log utils.Logger, services ...Configurable) error {
	clone, err := s.newInitializedConfig(newStoredConfig, log)
	if err != nil {
		return err
	}

	s.lock.Lock()
	s.conf = clone
	s.lock.Unlock()

	for _, service := range services {
		if err := service.Configure(*clone, log); err != nil {
			return errors.Wrapf(err, "failed to configure service %T", service)
		}
	}
	return nil
}

func (s *service) StoreConfig(sc StoredConfig, log utils.Logger) error {
	log.Debugf("Storing configuration, %v installed , %v listed apps, developer mode %v, allow http apps %v",
		len(sc.InstalledApps), len(sc.LocalManifests), sc.DeveloperModeOverride, sc.AllowHTTPAppsOverride)

	// Refresh computed values immediately, do not wait for OnConfigurationChanged
	err := s.Reconfigure(sc, log)
	if err != nil {
		return err
	}

	out := map[string]interface{}{}
	utils.Remarshal(&out, sc)

	// TODO test that SaveConfig will always cause OnConfigurationChange->c.Refresh
	return s.mm.Configuration.SavePluginConfig(out)
}

func (s *service) NewBaseLogger() utils.Logger {
	if pluginapi.IsConfiguredForDevelopment(s.mattermostConfig) {
		return utils.NewPluginLogger(s.mm, s)
	}
	return utils.NewPluginLogger(s.mm, nil)
}

func (s *service) GetLogConfig() utils.LogConfig {
	conf := s.Get()

	return utils.LogConfig{
		ChannelID:   conf.LogChannelID,
		Level:       zapcore.Level(conf.LogChannelLevel),
		BotUserID:   s.botUserID,
		IncludeJSON: conf.LogChannelJSON,
	}
}
