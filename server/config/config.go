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
	"github.com/mattermost/mattermost-plugin-apps/apps/awsapps"
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

	AWSLambdaAccessKey string
	AWSLambdaSecretKey string
	AWSS3Bucket        string
}

func (c Config) SetContextDefaults(cc *apps.Context) *apps.Context {
	if cc == nil {
		cc = &apps.Context{}
	}
	cc.BotUserID = c.BotUserID
	cc.MattermostSiteURL = c.MattermostSiteURL
	return cc
}

func (c Config) SetContextDefaultsForApp(appID apps.AppID, cc *apps.Context) *apps.Context {
	if cc == nil {
		cc = &apps.Context{}
	}
	cc = c.SetContextDefaults(cc)
	cc.AppID = appID
	cc.AppPath = path.Join(c.PluginURLPath, PathApps, string(appID))
	return cc
}

func (c Config) AppPath(appID apps.AppID) string {
	return c.PluginURL + PathApps + "/" + string(appID)
}

func (c *Config) Reconfigure(stored StoredConfig, mmconf *model.Config) error {
	mattermostSiteURL := mmconf.ServiceSettings.SiteURL
	if mattermostSiteURL == nil {
		return errors.New("plugin requires Mattermost Site URL to be set")
	}
	mattermostURL, err := url.Parse(*mattermostSiteURL)
	if err != nil {
		return err
	}

	c.StoredConfig = stored

	c.MattermostSiteURL = *mattermostSiteURL
	c.MattermostSiteHostname = mattermostURL.Hostname()
	c.PluginURLPath = "/plugins/" + c.BuildConfig.Manifest.Id
	c.PluginURL = strings.TrimRight(*mattermostSiteURL, "/") + c.PluginURLPath

	c.MaxWebhookSize = 75 * 1024 * 1024 // 75Mb
	if mmconf.FileSettings.MaxFileSize != nil {
		c.MaxWebhookSize = *mmconf.FileSettings.MaxFileSize
	}

	c.DeveloperMode = pluginapi.IsConfiguredForDevelopment(mmconf)

	// use CloudLambdaAccessEnvVar for now to detect the Cloud mode
	cloudLambdaAccessKey := os.Getenv(awsapps.CloudLambdaAccessEnvVar)
	c.MattermostCloudMode = false
	if cloudLambdaAccessKey != "" {
		c.MattermostCloudMode = true
	}

	if c.MattermostCloudMode {
		c.AWSLambdaAccessKey = os.Getenv(awsapps.CloudLambdaAccessEnvVar)
		c.AWSLambdaSecretKey = os.Getenv(awsapps.CloudLambdaSecretEnvVar)
		if c.AWSLambdaAccessKey == "" || c.AWSLambdaSecretKey == "" {
			return errors.Errorf("%s and %s must be set in cloud mode.", awsapps.CloudLambdaAccessEnvVar, awsapps.CloudLambdaSecretEnvVar)
		}
	} else {
		c.AWSLambdaAccessKey = os.Getenv(awsapps.LambdaAccessEnvVar)
		c.AWSLambdaSecretKey = os.Getenv(awsapps.LambdaSecretEnvVar)
	}
	c.AWSS3Bucket = awsapps.S3BucketName()

	return nil
}
