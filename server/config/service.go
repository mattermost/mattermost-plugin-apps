package config

import (
	"encoding/json"
	"sync"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/services/configservice"
)

type Configurable interface {
	Configure(Config)
}

// Configurator should be abbreviated as `cfg`
type Service interface {
	GetConfig() Config
	GetMattermostConfig() configservice.ConfigService
	Reconfigure(StoredConfig, ...Configurable) error
	StoreConfig(sc StoredConfig) error
}

var _ Service = (*service)(nil)

type service struct {
	BuildConfig
	botUserID string

	conf *Config

	lock             *sync.RWMutex
	mm               *pluginapi.Client
	mattermostConfig *model.Config
}

func NewService(mattermost *pluginapi.Client, buildConfig BuildConfig, botUserID string) Service {
	return &service{
		lock:        &sync.RWMutex{},
		mm:          mattermost,
		BuildConfig: buildConfig,
		botUserID:   botUserID,
	}
}

func (s *service) GetConfig() Config {
	s.lock.RLock()
	conf := s.conf
	s.lock.RUnlock()

	if conf == nil {
		return Config{
			BuildConfig: s.BuildConfig,
			BotUserID:   s.botUserID,
		}
	}
	return *conf
}

func (s *service) GetMattermostConfig() configservice.ConfigService {
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

	newConfig := s.GetConfig()
	err := newConfig.Reconfigure(stored, mmconf)
	if err != nil {
		return err
	}

	s.lock.Lock()
	s.conf = &newConfig
	s.lock.Unlock()

	for _, s := range services {
		s.Configure(newConfig)
	}

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
