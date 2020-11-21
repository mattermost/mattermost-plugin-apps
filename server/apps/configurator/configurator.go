package configurator

import (
	"net/url"
	"strings"
	"sync"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
)

var _ apps.Configurator = (*config)(nil)

type config struct {
	*apps.BuildConfig
	botUserID string

	conf *apps.Config

	lock             *sync.RWMutex
	mm               *pluginapi.Client
	mattermostConfig *model.Config
}

func NewConfigurator(mattermost *pluginapi.Client, buildConfig *apps.BuildConfig, botUserID string) apps.Configurator {
	return &config{
		lock:        &sync.RWMutex{},
		mm:          mattermost,
		BuildConfig: buildConfig,
		botUserID:   botUserID,
	}
}

func (c *config) GetConfig() apps.Config {
	c.lock.RLock()
	conf := c.conf
	c.lock.RUnlock()

	if conf == nil {
		return apps.Config{
			BuildConfig: c.BuildConfig,
			BotUserID:   c.botUserID,
		}
	}
	return *conf
}

func (c *config) GetMattermostConfig() *model.Config {
	c.lock.RLock()
	mmconf := c.mattermostConfig
	c.lock.RUnlock()

	if mmconf == nil {
		mmconf = c.mm.Configuration.GetConfig()
		c.lock.Lock()
		c.mattermostConfig = mmconf
		c.lock.Unlock()
	}
	return c.mattermostConfig
}

func (c *config) RefreshConfig(stored *apps.StoredConfig) error {
	mattermostSiteURL := c.GetMattermostConfig().ServiceSettings.SiteURL
	if mattermostSiteURL == nil {
		return errors.New("plugin requires Mattermost Site URL to be set")
	}
	mattermostURL, err := url.Parse(*mattermostSiteURL)
	if err != nil {
		return err
	}
	pluginURLPath := "/plugins/" + c.BuildConfig.Manifest.Id
	pluginURL := strings.TrimRight(*mattermostSiteURL, "/") + pluginURLPath

	newConfig := c.GetConfig()
	newConfig.StoredConfig = stored
	newConfig.MattermostSiteURL = *mattermostSiteURL
	newConfig.MattermostSiteHostname = mattermostURL.Hostname()
	newConfig.PluginURL = pluginURL
	newConfig.PluginURLPath = pluginURLPath

	c.lock.Lock()
	c.conf = &newConfig
	c.lock.Unlock()

	return nil
}

func (c *config) StoreConfig(newStored apps.ConfigMapper) error {
	// TODO test that SaveConfig will always cause OnConfigurationChange->c.Refresh
	return c.mm.Configuration.SavePluginConfig(newStored.ConfigMap())
}
