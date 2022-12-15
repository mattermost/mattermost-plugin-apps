package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"strings"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	appspath "github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// StoredConfig represents the data stored in and managed with the Mattermost
// config.
//
// StoredConfig should be abbreviated as sc.
type StoredConfig struct {
	// InstalledApps is a list of all apps installed on the Mattermost instance.
	//
	// For each installed app, an entry of string(AppID) -> sha1(App) is added,
	// and the App struct is stored in KV under app_<sha1(App)>. Implementation
	// in `store.App`.
	InstalledApps map[string]string `json:"installed_apps,omitempty"`

	// LocalManifests is a list of locally-stored manifests. Local is in
	// contrast to the "global" list of manifests which in the initial version
	// is loaded from S3.
	//
	// For each installed app, an entry of string(AppID) -> sha1(Manifest) is
	// added, and the Manifest struct is stored in KV under
	// manifest_<sha1(Manifest)>. Implementation in `store.Manifest`.
	LocalManifests map[string]string `json:"local_manifests,omitempty"`

	DeveloperModeOverride *bool `json:"developer_mode"`
	AllowHTTPAppsOverride *bool `json:"allow_http_apps"`

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

func (s *service) newConfig() Config {
	return Config{
		PluginManifest: s.pluginManifest,
		BuildDate:      BuildDate,
		BuildHash:      BuildHash,
		BuildHashShort: BuildHashShort,
		BotUserID:      s.botUserID,
	}
}

func (s *service) newInitializedConfig(newStoredConfig StoredConfig, log utils.Logger) (*Config, error) {
	conf := s.newConfig()
	conf.StoredConfig = newStoredConfig
	newMattermostConfig := s.reloadMattermostConfig()

	if conf.DeveloperModeOverride != nil {
		conf.DeveloperMode = *conf.DeveloperModeOverride
	} else {
		conf.DeveloperMode = pluginapi.IsConfiguredForDevelopment(newMattermostConfig)
	}

	mattermostSiteURL := newMattermostConfig.ServiceSettings.SiteURL
	if mattermostSiteURL == nil {
		return nil, errors.New("plugin requires Mattermost Site URL to be set")
	}
	u, err := url.Parse(*mattermostSiteURL)
	if err != nil {
		return nil, errors.Wrap(err, "invalid SiteURL in config")
	}

	var localURL string
	if newMattermostConfig.ServiceSettings.ConnectionSecurity != nil && *newMattermostConfig.ServiceSettings.ConnectionSecurity == model.ConnSecurityTLS {
		// If there is no reverse proxy use the server URL
		localURL = u.String()
	} else {
		// Avoid the reverse proxy by using the local port
		listenAddress := newMattermostConfig.ServiceSettings.ListenAddress
		if listenAddress == nil {
			return nil, errors.New("plugin requires Mattermost Listen Address to be set")
		}
		host, port, err := net.SplitHostPort(*listenAddress)
		if err != nil {
			return nil, err
		}

		if host == "" {
			host = "127.0.0.1"
		}

		localURL = "http://" + host + ":" + port + u.Path
	}

	conf.MattermostSiteURL = u.String()
	conf.MattermostLocalURL = localURL
	conf.PluginURLPath = "/plugins/" + conf.PluginManifest.Id
	conf.PluginURL = strings.TrimRight(u.String(), "/") + conf.PluginURLPath

	conf.MaxWebhookSize = 75 * 1024 * 1024 // 75Mb
	if newMattermostConfig.FileSettings.MaxFileSize != nil {
		conf.MaxWebhookSize = int(*newMattermostConfig.FileSettings.MaxFileSize)
	}

	conf.AWSAccessKey = os.Getenv(upaws.AccessEnvVar)
	conf.AWSSecretKey = os.Getenv(upaws.SecretEnvVar)
	conf.AWSRegion = upaws.Region()
	conf.AWSS3Bucket = upaws.S3BucketName()

	license := s.getMattermostLicense(log)
	conf.MattermostCloudMode = license != nil &&
		license.Features != nil &&
		license.Features.Cloud != nil &&
		*license.Features.Cloud
	if conf.MattermostCloudMode {
		log.Debugf("Detected Mattermost Cloud mode based on the license")
	}

	// On community.mattermost.com license is not suitable for checking, resort
	// to the presence of legacy environment variable to trigger it.
	legacyAccessKey := os.Getenv(upaws.DeprecatedCloudAccessEnvVar)
	if legacyAccessKey != "" {
		conf.MattermostCloudMode = true
		log.Debugf("Detected Mattermost Cloud mode based on the %s variable", upaws.DeprecatedCloudAccessEnvVar)
		conf.AWSAccessKey = legacyAccessKey
	}

	if conf.MattermostCloudMode {
		legacySecretKey := os.Getenv(upaws.DeprecatedCloudSecretEnvVar)
		if legacySecretKey != "" {
			conf.AWSSecretKey = legacySecretKey
		}
		if conf.AWSAccessKey == "" || conf.AWSSecretKey == "" {
			return nil, errors.New("access credentials for AWS must be set in Mattermost Cloud mode")
		}
	}

	if conf.AllowHTTPAppsOverride != nil {
		conf.AllowHTTPApps = *conf.AllowHTTPAppsOverride
	} else {
		conf.AllowHTTPApps = !conf.MattermostCloudMode || conf.DeveloperMode
	}

	return &conf, nil
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
