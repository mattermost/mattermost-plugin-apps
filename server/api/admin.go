package api

import (
	"github.com/mattermost/mattermost-plugin-apps/modelapps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type Admin interface {
	ListApps() ([]*modelapps.App, md.MD, error)
	InstallApp(*modelapps.Context, modelapps.SessionToken, *modelapps.InInstallApp) (*modelapps.App, md.MD, error)
	ProvisionApp(*modelapps.Context, modelapps.SessionToken, *modelapps.InProvisionApp) (*modelapps.App, md.MD, error)
}
