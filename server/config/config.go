package config

import (
	"path"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
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

	DeveloperMode bool

	BotUserID              string
	MattermostSiteHostname string
	MattermostSiteURL      string
	PluginURL              string
	PluginURLPath          string

	// Maximum size of incoming remote webhook messages
	MaxWebhookSize int64
}

func (conf Config) SetContextDefaults(cc *apps.Context) *apps.Context {
	if cc == nil {
		cc = &apps.Context{}
	}
	cc.BotUserID = conf.BotUserID
	cc.MattermostSiteURL = conf.MattermostSiteURL
	return cc
}

func (conf Config) SetContextDefaultsForApp(appID apps.AppID, cc *apps.Context) *apps.Context {
	if cc == nil {
		cc = &apps.Context{}
	}
	cc = conf.SetContextDefaults(cc)
	cc.AppID = appID
	cc.AppPath = path.Join(conf.PluginURLPath, PathApps, string(appID))
	return cc
}

func (conf Config) AppPath(appID apps.AppID) string {
	return conf.PluginURL + path.Join(PathApps, string(appID))
}

func (conf Config) StaticPath(appID apps.AppID, name string) string {
	return path.Join(conf.AppPath(appID), apps.StaticFolder, name)
}
