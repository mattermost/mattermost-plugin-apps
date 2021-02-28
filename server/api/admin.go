package api

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type Admin interface {
	ListApps() (map[apps.AppID]*apps.App, md.MD, error)
	GetApp(appID apps.AppID) (*apps.App, error)
	GetManifest(appID apps.AppID) (*apps.Manifest, error)
	InstallApp(*apps.Context, apps.SessionToken, *apps.InInstallApp) (*apps.App, md.MD, error)
	UninstallApp(appID apps.AppID) error
	InstallManifest(*apps.Context, apps.SessionToken, *apps.Manifest) (md.MD, error)
	// LoadAppsList() error
}
