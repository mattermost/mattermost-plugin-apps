package api

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type Admin interface {
	AddLocalManifest(*apps.Context, apps.SessionToken, *apps.Manifest) (md.MD, error)
	GetApp(appID apps.AppID) (*apps.App, error)
	GetManifest(appID apps.AppID) (*apps.Manifest, error)
	InstallApp(*apps.Context, apps.SessionToken, *apps.InInstallApp) (*apps.App, md.MD, error)
	ListInstalledApps() map[apps.AppID]*apps.App
	ListMarketplaceApps(filter string) map[apps.AppID]*apps.MarketplaceApp
	UninstallApp(appID apps.AppID) error
	// LoadAppsList() error
}
