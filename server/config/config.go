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

	BotUserID              string
	MattermostSiteHostname string
	MattermostSiteURL      string
	PluginURL              string
	PluginURLPath          string
}

func (c Config) NewContext() *apps.Context {
	return &apps.Context{
		BotUserID:         c.BotUserID,
		MattermostSiteURL: c.MattermostSiteURL,
	}
}

func (c Config) NewContextForApp(appID apps.AppID) *apps.Context {
	return &apps.Context{
		AppID:             appID,
		BotUserID:         c.BotUserID,
		MattermostSiteURL: c.MattermostSiteURL,
		AppPath:           path.Join(c.PluginURLPath, string(appID)),
	}
}
