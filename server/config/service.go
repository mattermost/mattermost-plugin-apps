package config

import (
	"encoding/json"
	"sync"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/i18n"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/services/configservice"

	"github.com/mattermost/mattermost-plugin-apps/server/telemetry"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Configurable interface {
	Configure(Config) error
}

// Configurator should be abbreviated as `cfg`
type Service interface {
	Basic() (Config, *pluginapi.Client, utils.Logger)
	Get() Config
	Logger() utils.Logger
	MattermostAPI() *pluginapi.Client
	MattermostConfig() configservice.ConfigService
	I18N() *i18n.Bundle
	Telemetry() *telemetry.Telemetry

	Reconfigure(StoredConfig, ...Configurable) error
	StoreConfig(sc StoredConfig) error
}

var _ Service = (*service)(nil)

type service struct {
	pluginManifest model.Manifest
	botUserID      string
	log            utils.Logger
	mm             *pluginapi.Client
	i18n           *i18n.Bundle
	telemetry      *telemetry.Telemetry

	lock             *sync.RWMutex
	conf             *Config
	mattermostConfig *model.Config
}

func NewService(mm *pluginapi.Client, pliginManifest model.Manifest, botUserID string, telemetry *telemetry.Telemetry, i18nBundle *i18n.Bundle) Service {
	// utils.NewPluginLogger(mm).Debugf("<>/<> NewService 1 %q", BuildHash)
	return &service{
		pluginManifest: pliginManifest,
		botUserID:      botUserID,
		log:            utils.NewPluginLogger(mm),
		mm:             mm,
		lock:           &sync.RWMutex{},
		i18n:           i18nBundle,
		telemetry:      telemetry,
	}
}

// Basic is a convenience method, included in the interface so one can write:
//   conf, mm, log := x.conf.Basic()
func (s *service) Basic() (Config, *pluginapi.Client, utils.Logger) {
	return s.Get(),
		s.MattermostAPI(),
		s.Logger()
}

func (s *service) Get() Config {
	s.lock.RLock()
	conf := s.conf
	s.lock.RUnlock()

	if conf == nil {
		return Config{
			PluginManifest: s.pluginManifest,
			BuildDate:      BuildDate,
			BuildHash:      BuildHash,
			BuildHashShort: BuildHashShort,
			BotUserID:      s.botUserID,
		}
	}
	return *conf
}

func (s *service) MattermostAPI() *pluginapi.Client {
	return s.mm
}

func (s *service) Logger() utils.Logger {
	return s.log
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

func (s *service) Reconfigure(stored StoredConfig, services ...Configurable) error {
	mmconf := s.reloadMattermostConfig()
	newConfig := s.Get()
	s.log.Debugf("<>/<> 0 %q %q", BuildHash, newConfig.BuildHash)

	// GetLicense silently drops an RPC error
	// (https://github.com/mattermost/mattermost-server/blob/fc75b72bbabf7fabfad24b9e1e4c321ca9b9b7f1/plugin/client_rpc_generated.go#L864).
	// When running in Mattermost cloud we must not fall back to the on-prem mode, so in case we get a nil retry once.
	license := s.mm.System.GetLicense()
	if license == nil {
		license = s.mm.System.GetLicense()
		if license == nil {
			s.log.Infof("Failed to fetch license twice. May incorrectly default to on-prem mode.")
		}
	}

	s.log.Debugf("<>/<> 1 %q %q", BuildHash, newConfig.BuildHash)
	err := newConfig.Update(stored, mmconf, license, s.log)
	if err != nil {
		return err
	}
	s.log.Debugf("<>/<> 2 %q %q", BuildHash, newConfig.BuildHash)

	s.lock.Lock()
	s.conf = &newConfig
	s.lock.Unlock()

	for _, s := range services {
		err = s.Configure(newConfig)
		if err != nil {
			return errors.Wrapf(err, "error configuring %T", s)
		}
	}

	s.log.Debugf("<>/<> 3 %q %q", BuildHash, newConfig.BuildHash)
	return nil
}

func (s *service) StoreConfig(sc StoredConfig) error {
	// Refresh computed values immediately, do not wait for OnConfigurationChanged
	err := s.Reconfigure(sc)
	if err != nil {
		return err
	}

	data, err := json.Marshal(sc)
	if err != nil {
		return err
	}
	out := map[string]interface{}{}
	err = json.Unmarshal(data, &out)
	if err != nil {
		return err
	}

	// TODO test that SaveConfig will always cause OnConfigurationChange->c.Refresh
	return s.mm.Configuration.SavePluginConfig(out)
}
