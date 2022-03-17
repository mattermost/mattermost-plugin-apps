package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"regexp"
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

	DeveloperMode       bool
	AllowHTTPApps       bool
	MattermostCloudMode bool

	BotUserID          string
	MattermostSiteURL  string
	MattermostLocalURL string
	PluginURL          string
	PluginURLPath      string

	// Maximum size of incoming remote webhook messages
	MaxWebhookSize int64

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

// developerModeDomains is the list of domains for which DevelopmentMode will be
// forced on. Empty for the time being.
var developerModeDomains = regexp.MustCompile("^" + strings.Join([]string{}, "|") + "$")

// allowHTTPAppsDomains is the list of domains for which AllowHTTPApps will be
// forced on. 
var allowHTTPAppsDomains = regexp.MustCompile("^" + strings.Join([]string{
	`.*\.test\.mattermost\.cloud`,
	`community\.mattermost\.com`,
	`community-[a-z]+\.mattermost\.com`,
}, "|") + "$")

func (conf *Config) Update(stored StoredConfig, mmconf *model.Config, license *model.License, log utils.Logger) error {
	mattermostSiteURL := mmconf.ServiceSettings.SiteURL
	if mattermostSiteURL == nil {
		return errors.New("plugin requires Mattermost Site URL to be set")
	}
	u, err := url.Parse(*mattermostSiteURL)
	if err != nil {
		return err
	}

	var localURL string
	if mmconf.ServiceSettings.ConnectionSecurity != nil && *mmconf.ServiceSettings.ConnectionSecurity == model.ConnSecurityTLS {
		// If there is no reverse proxy use the server URL
		localURL = u.String()
	} else {
		// Avoid the reverse proxy by using the local port
		listenAddress := mmconf.ServiceSettings.ListenAddress
		if listenAddress == nil {
			return errors.New("plugin requires Mattermost Listen Address to be set")
		}
		host, port, err := net.SplitHostPort(*listenAddress)
		if err != nil {
			return err
		}

		if host == "" {
			host = "127.0.0.1"
		}

		localURL = "http://" + host + ":" + port + u.Path
	}

	conf.StoredConfig = stored

	conf.MattermostSiteURL = u.String()
	conf.MattermostLocalURL = localURL
	conf.PluginURLPath = "/plugins/" + conf.PluginManifest.Id
	conf.PluginURL = strings.TrimRight(u.String(), "/") + conf.PluginURLPath

	conf.MaxWebhookSize = 75 * 1024 * 1024 // 75Mb
	if mmconf.FileSettings.MaxFileSize != nil {
		conf.MaxWebhookSize = *mmconf.FileSettings.MaxFileSize
	}

	conf.DeveloperMode = pluginapi.IsConfiguredForDevelopment(mmconf)
	if developerModeDomains.MatchString(u.Hostname()) {
		conf.DeveloperMode = true
	}

	conf.AllowHTTPApps = !conf.MattermostCloudMode || conf.DeveloperMode
	if allowHTTPAppsDomains.MatchString(u.Hostname()) {
		conf.AllowHTTPApps = true
	}

	conf.AWSAccessKey = os.Getenv(upaws.AccessEnvVar)
	conf.AWSSecretKey = os.Getenv(upaws.SecretEnvVar)
	conf.AWSRegion = upaws.Region()
	conf.AWSS3Bucket = upaws.S3BucketName()

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
			return errors.New("access credentials for AWS must be set in Mattermost Cloud mode")
		}
	}

	return nil
}

func (conf Config) GetPluginVersionInfo() map[string]interface{} {
	return map[string]interface{}{
		"version": conf.PluginManifest.Version,
	}
}

func (c *Config) InfoTemplateData() map[string]string {
	return map[string]string{
		"Version":       c.PluginManifest.Version,
		"URL":           fmt.Sprintf("[%s](https://github.com/mattermost/%s/commit/%s)", c.BuildHashShort, Repository, c.BuildHash),
		"BuildDate":     c.BuildDate,
		"CloudMode":     fmt.Sprintf("%t", c.MattermostCloudMode),
		"DeveloperMode": fmt.Sprintf("%t", c.DeveloperMode),
		"AllowHTTPApps": fmt.Sprintf("%t", c.AllowHTTPApps),
	}
}
