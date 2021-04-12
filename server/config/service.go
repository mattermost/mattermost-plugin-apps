package config

import (
	"encoding/json"
	"net/url"
	"strings"
	"sync"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
)

type Configurable interface {
	Configure(Config)
}

// Configurator should be abbreviated as `cfg`
type Service interface {
	GetConfig() Config
	GetMattermostConfig() *model.Config
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

func (s *service) GetMattermostConfig() *model.Config {
	s.lock.RLock()
	mmconf := s.mattermostConfig
	s.lock.RUnlock()

	if mmconf == nil {
		mmconf = s.mm.Configuration.GetConfig()
		s.lock.Lock()
		s.mattermostConfig = mmconf
		s.lock.Unlock()
	}
	return mmconf
}

func (s *service) Reconfigure(stored StoredConfig, services ...Configurable) error {
	mmconf := s.GetMattermostConfig()
	mattermostSiteURL := mmconf.ServiceSettings.SiteURL
	if mattermostSiteURL == nil {
		return errors.New("plugin requires Mattermost Site URL to be set")
	}
	mattermostURL, err := url.Parse(*mattermostSiteURL)
	if err != nil {
		return err
	}
	pluginURLPath := "/plugins/" + s.BuildConfig.Manifest.Id
	pluginURL := strings.TrimRight(*mattermostSiteURL, "/") + pluginURLPath

	newConfig := s.GetConfig()
	newConfig.StoredConfig = stored

	newConfig.MattermostSiteURL = *mattermostSiteURL
	newConfig.MattermostSiteHostname = mattermostURL.Hostname()
	newConfig.PluginURL = pluginURL
	newConfig.PluginURLPath = pluginURLPath

	newConfig.MaxWebhookSize = 75 * 1024 * 1024 // 75Mb
	if mmconf.FileSettings.MaxFileSize != nil {
		newConfig.MaxWebhookSize = *mmconf.FileSettings.MaxFileSize
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
	if strings.HasPrefix(*s.mattermostConfig.ServiceSettings.SiteURL, "http://localhost:") {
		return nil
	}
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
