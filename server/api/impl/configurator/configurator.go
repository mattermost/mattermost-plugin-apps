package configurator

import (
	"net/url"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/mattermost/mattermost-plugin-apps/server/api"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
)

var _ api.Configurator = (*config)(nil)

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
	return c.mattermostConfig
}

func (c *config) RefreshConfig(stored *api.StoredConfig) error {
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

	prevStored := newConfig.StoredConfig
	if prevStored != nil {
		if prevStored.AWSSecretAccessKey != stored.AWSSecretAccessKey ||
			prevStored.AWSAccessKeyID != stored.AWSAccessKeyID {
			newConfig.AWSSession, err = session.NewSession(&aws.Config{
				Region:      aws.String("us-east-2"),
				Credentials: credentials.NewStaticCredentials(stored.AWSAccessKeyID, stored.AWSSecretAccessKey, ""),
			})
			if err != nil {
				return err
			}
		}
	}

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

func (c *config) StoreConfig(newStored api.ConfigMapper) error {
	// TODO test that SaveConfig will always cause OnConfigurationChange->c.Refresh
	return c.mm.Configuration.SavePluginConfig(newStored.ConfigMap())
}
