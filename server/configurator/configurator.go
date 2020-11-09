package configurator

import (
	"net/url"
	"strings"
	"sync"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
)

type ConfigMapper interface {
	ConfigMap() (result map[string]interface{})
}

type Service interface {
	GetConfig() Config
	GetMattermostConfig() *model.Config
	Refresh(*StoredConfig) error
	Store(ConfigMapper) error
}

var _ Service = (*service)(nil)

type service struct {
	*BuildConfig
	botUserID string

	conf *Config

	lock             *sync.RWMutex
	mattermost       *pluginapi.Client
	mattermostConfig *model.Config
}

func NewConfigurator(mattermost *pluginapi.Client, buildConfig *BuildConfig, botUserID string) Service {
	return &service{
		lock:        &sync.RWMutex{},
		mattermost:  mattermost,
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
		mmconf = s.mattermost.Configuration.GetConfig()
		s.lock.Lock()
		s.mattermostConfig = mmconf
		s.lock.Unlock()
	}
	return s.mattermostConfig
}

func (s *service) Refresh(stored *StoredConfig) error {
	mattermostSiteURL := s.GetMattermostConfig().ServiceSettings.SiteURL
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

	s.lock.Lock()
	s.conf = &newConfig
	s.lock.Unlock()

	return nil
}

func (s *service) Store(newStored ConfigMapper) error {
	// TODO test that SaveConfig will always cause OnConfigurationChange->c.Refresh
	return s.mattermost.Configuration.SavePluginConfig(newStored.ConfigMap())
}
