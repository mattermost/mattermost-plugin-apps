package config

import (
	"net/url"
	"os"
	"path"
	"strings"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
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
}

type BuildConfig struct {
	model.Manifest
	BuildDate      string
	BuildHash      string
	BuildHashShort string
}

// Config represents the the metadata handed to all request runners (command,
// http).
//
// Config should be abbreviated as `conf`.
type Config struct {
	StoredConfig
	BuildConfig

	DeveloperMode       bool
	MattermostCloudMode bool

	BotUserID              string
	MattermostSiteHostname string
	MattermostSiteURL      string
	PluginURL              string
	PluginURLPath          string

	// Maximum size of incoming remote webhook messages
	MaxWebhookSize int64

	AWSRegion    string
	AWSAccessKey string
	AWSSecretKey string
	AWSS3Bucket  string
}

func (conf Config) AppURL(appID apps.AppID) string {
	return conf.PluginURL + path.Join(PathApps, string(appID))
}

// StaticURL returns the URL to a static asset.
func (conf Config) StaticURL(appID apps.AppID, name string) string {
	return conf.AppURL(appID) + "/" + path.Join(apps.StaticFolder, name)
}

func (conf *Config) Reconfigure(stored StoredConfig, mmconf *model.Config, license *model.License) error {
	mattermostSiteURL := mmconf.ServiceSettings.SiteURL
	if mattermostSiteURL == nil {
		return errors.New("plugin requires Mattermost Site URL to be set")
	}
	mattermostURL, err := url.Parse(*mattermostSiteURL)
	if err != nil {
		return err
	}

	conf.StoredConfig = stored

	conf.MattermostSiteURL = *mattermostSiteURL
	conf.MattermostSiteHostname = mattermostURL.Hostname()
	conf.PluginURLPath = "/plugins/" + conf.BuildConfig.Manifest.Id
	conf.PluginURL = strings.TrimRight(*mattermostSiteURL, "/") + conf.PluginURLPath

	conf.MaxWebhookSize = 75 * 1024 * 1024 // 75Mb
	if mmconf.FileSettings.MaxFileSize != nil {
		conf.MaxWebhookSize = *mmconf.FileSettings.MaxFileSize
	}

	conf.DeveloperMode = pluginapi.IsConfiguredForDevelopment(mmconf)

	conf.AWSAccessKey = os.Getenv(upaws.AccessEnvVar)
	conf.AWSSecretKey = os.Getenv(upaws.SecretEnvVar)
	conf.AWSRegion = upaws.Region()
	conf.AWSS3Bucket = upaws.S3BucketName()

	conf.MattermostCloudMode = license != nil &&
		license.Features != nil &&
		license.Features.Cloud != nil &&
		*license.Features.Cloud

	// On community.mattermost.com license is not suitable for checking, resort
	// to the presence of legacy environment variable to trigger it.
	legacyAccessKey := os.Getenv(upaws.DeprecatedCloudAccessEnvVar)
	if legacyAccessKey != "" {
		conf.MattermostCloudMode = true
		conf.AWSAccessKey = legacyAccessKey
	}

	if conf.MattermostCloudMode {
		legacySecretKey := os.Getenv(upaws.DeprecatedCloudSecretEnvVar)
		if legacySecretKey != "" {
			conf.AWSSecretKey = legacySecretKey
		}
		if conf.AWSAccessKey == "" || conf.AWSSecretKey == "" {
			return errors.New("access credentials for AWS must be set in Mattermost Cloud mode")
		}
	}

	return nil
}
