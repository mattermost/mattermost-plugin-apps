package api

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type Admin interface {
	AddLocalManifest(*apps.Context, apps.SessionToken, *apps.Manifest) (md.MD, error)
	GetInstalledApp(appID apps.AppID) (*apps.App, error)
	GetInstalledApps() []*apps.App
	GetListedApps(filter string) []*apps.ListedApp
	GetManifest(appID apps.AppID) (*apps.Manifest, error)
	InstallApp(*apps.Context, apps.SessionToken, *apps.InInstallApp) (*apps.App, md.MD, error)
	SynchronizeInstalledApps() error
	UninstallApp(appID apps.AppID) error
}
