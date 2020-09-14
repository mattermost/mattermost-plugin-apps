package configurator

import (
	"errors"
	"net/url"
	"strings"
	"sync"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
)

type Mapper interface {
	MapOnto(onto map[string]interface{}) (result map[string]interface{})
}

type Configurator interface {
	Get() Config
	GetMattermostConfig() *model.Config
	Refresh() error
	Store(Mapper) error
}

var _ Configurator = (*configurator)(nil)

type configurator struct {
	*BuildConfig
	conf *Config

	lock   *sync.RWMutex
	mm     *pluginapi.Client
	mmconf *model.Config
}

func NewConfigurator(buildConfig *BuildConfig, mm *pluginapi.Client) Configurator {
	return &configurator{
		lock:        &sync.RWMutex{},
		mm:          mm,
		BuildConfig: buildConfig,
	}
}

func (c *configurator) Get() Config {
	c.lock.RLock()
	conf := c.conf
	c.lock.RUnlock()

	if conf == nil {
		return Config{}
	}
	return *conf
}

func (c *configurator) GetMattermostConfig() *model.Config {
	c.lock.RLock()
	mmconf := c.mmconf
	c.lock.RUnlock()

	if mmconf == nil {
		mmconf = c.mm.Configuration.GetConfig()
		c.lock.Lock()
		c.mmconf = mmconf
		c.lock.Unlock()
	}
	return c.mmconf
}

func (c *configurator) Refresh() error {
	stored := StoredConfig{}
	_ = c.mm.Configuration.LoadPluginConfiguration(&stored)

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

	newConfig := c.conf
	if newConfig == nil {
		newConfig = &Config{}
	}
	newConfig.StoredConfig = &stored
	newConfig.BuildConfig = c.BuildConfig
	newConfig.MattermostSiteURL = *mattermostSiteURL
	newConfig.MattermostSiteHostname = mattermostURL.Hostname()
	newConfig.PluginURL = pluginURL
	newConfig.PluginURLPath = pluginURLPath

	c.lock.Lock()
	c.conf = newConfig
	c.lock.Unlock()

	return nil
}

func (c *configurator) Store(newStored Mapper) error {
	// TODO test that SaveConfig will always cause OnConfigurationChange->c.Refresh
	return c.mm.Configuration.SavePluginConfig(newStored.MapOnto(nil))
}
