package api

import (
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// StoredConfig represents the data stored in and managed with the Mattermost
// config.
type StoredConfig struct {
	Apps               map[apps.AppID]interface{}
	AWSAccessKeyID     string
	AWSSecretAccessKey string
}

type ConfigMapper interface {
	ConfigMap() (result map[string]interface{})
}

func (sc *StoredConfig) ConfigMap() map[string]interface{} {
	return map[string]interface{}{
		"Apps":               sc.Apps,
		"AWSAccessKeyID":     sc.AWSAccessKeyID,
		"AWSSecretAccessKey": sc.AWSSecretAccessKey,
	}
}

type BuildConfig struct {
	*model.Manifest
	BuildDate      string
	BuildHash      string
	BuildHashShort string
}

// Config represents the the metadata handed to all request runners (command,
// http).
type Config struct {
	*StoredConfig
	*BuildConfig

	BotUserID              string
	MattermostSiteHostname string
	MattermostSiteURL      string
	PluginURL              string
	PluginURLPath          string
}

type Configurator interface {
	GetConfig() Config
	GetMattermostConfig() *model.Config
	RefreshConfig(*StoredConfig) error
	StoreConfig(ConfigMapper) error
}
