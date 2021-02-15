package api

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type Admin interface {
	ListApps() ([]*apps.App, md.MD, error)
	GetApp(appID apps.AppID) (*apps.App, error)
	InstallApp(*apps.Context, apps.SessionToken, *apps.InInstallApp) (*apps.App, md.MD, error)
	ProvisionApp(*apps.Context, apps.SessionToken, *apps.InProvisionApp) (*apps.App, md.MD, error)
	LoadAppsList() error
}
