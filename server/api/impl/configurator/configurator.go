package configurator

import (
	"encoding/json"
	"net/url"
	"strings"
	"sync"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

type config struct {
	*api.BuildConfig
	botUserID string

	conf *api.Config

	lock             *sync.RWMutex
	mm               *pluginapi.Client
	mattermostConfig *model.Config
}

func NewConfigurator(mattermost *pluginapi.Client, buildConfig *api.BuildConfig, botUserID string) api.Configurator {
	return &config{
		lock:        &sync.RWMutex{},
		mm:          mattermost,
		BuildConfig: buildConfig,
		botUserID:   botUserID,
	}
}

func (c *config) GetConfig() api.Config {
	c.lock.RLock()
	conf := c.conf
	c.lock.RUnlock()

	if conf == nil {
		return api.Config{
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
	return mmconf
}

func (c *config) Reconfigure(stored *api.StoredConfig, services ...api.Configurable) error {
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

	for _, s := range services {
		err = s.Configure(newConfig)
		if err != nil {
			return errors.Wrapf(err, "failed to reconfigure service %T", s)
		}
	}

	return nil
}

func (c *config) StoreConfig(sc *api.StoredConfig) error {
	// Reconfigure computed values immediately, do not wait for OnConfigurationChanged
	err := c.Reconfigure(sc)
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

	// TODO test that SaveConfig will always cause OnConfigurationChange->c.Reconfigure
	return c.mm.Configuration.SavePluginConfig(out)
}
