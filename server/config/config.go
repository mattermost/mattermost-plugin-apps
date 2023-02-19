package config

import (
	"fmt"
	"path"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	appspath "github.com/mattermost/mattermost-plugin-apps/apps/path"
)

// StoredConfig represents the data stored in and managed with the Mattermost
// config.
//
// StoredConfig should be abbreviated as sc.
type StoredConfig struct {
	DeveloperModeOverride *bool `json:"developer_mode"`
	AllowHTTPAppsOverride *bool `json:"allow_http_apps"`

	CachedStore string `json:"cache_store,omitempty"`

	LogChannelID    string `json:"log_channel_id,omitempty"`
	LogChannelLevel int    `json:"log_channel_level,omitempty"`
	LogChannelJSON  bool   `json:"log_channel_json,omitempty"`
}

var BuildDate string
var BuildHash string
var BuildHashShort string

// Config represents the the metadata handed to all request runners (command,
// http).
//
// Config should be abbreviated as `conf`.
type Config struct {
	StoredConfig

	PluginManifest model.Manifest
	BuildDate      string
	BuildHash      string
	BuildHashShort string

	MattermostCloudMode bool
	DeveloperMode       bool
	AllowHTTPApps       bool

	BotUserID          string
	MattermostSiteURL  string
	MattermostLocalURL string
	PluginURL          string
	PluginURLPath      string

	// Maximum size of incoming remote webhook messages
	MaxWebhookSize int

	AWSRegion    string
	AWSAccessKey string
	AWSSecretKey string
	AWSS3Bucket  string
}

func (conf Config) AppURL(appID apps.AppID) string {
	return conf.PluginURL + path.Join(appspath.Apps, string(appID))
}

// StaticURL returns the URL to a static asset.
func (conf Config) StaticURL(appID apps.AppID, name string) string {
	return conf.AppURL(appID) + "/" + path.Join(appspath.StaticFolder, name)
}

func (conf Config) GetPluginVersionInfo() map[string]interface{} {
	return map[string]interface{}{
		"version": conf.PluginManifest.Version,
	}
}

func (conf *Config) InfoTemplateData() map[string]string {
	return map[string]string{
		"Version":       conf.PluginManifest.Version,
		"URL":           fmt.Sprintf("[%s](https://github.com/mattermost/%s/commit/%s)", conf.BuildHashShort, Repository, conf.BuildHash),
		"BuildDate":     conf.BuildDate,
		"CloudMode":     fmt.Sprintf("%t", conf.MattermostCloudMode),
		"DeveloperMode": fmt.Sprintf("%t", conf.DeveloperMode),
		"AllowHTTPApps": fmt.Sprintf("%t", conf.AllowHTTPApps),
	}
}

func (conf Config) Loggable() []interface{} {
	return append([]interface{}{},
		"version", conf.PluginManifest.Version,
		"commit", conf.BuildHashShort,
		"build_date", conf.BuildDate,
		"cloud_mode", fmt.Sprintf("%t", conf.MattermostCloudMode),
		"developer_mode", fmt.Sprintf("%t", conf.DeveloperMode),
		"allow_http_apps", fmt.Sprintf("%t", conf.AllowHTTPApps),
	)
}
