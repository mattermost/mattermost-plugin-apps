package store

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

type appStoreMock struct {
	apps map[apps.AppID]*apps.App
}

var _ AppStore = (*appStoreMock)(nil)

func NewAppStoreMock(apps map[apps.AppID]*apps.App) AppStore {
	return &appStoreMock{apps: apps}
}

func (a *appStoreMock) AsMap() map[apps.AppID]*apps.App {
	return a.apps
}

func (a *appStoreMock) Configure(c config.Config) {}

func (a *appStoreMock) Delete(appID apps.AppID) error {
	if _, ok := a.apps[appID]; !ok {
		return utils.ErrNotFound
	}
	delete(a.apps, appID)
	return nil
}

func (a *appStoreMock) Get(appID apps.AppID) (*apps.App, error) {
	app, ok := a.apps[appID]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return app, nil
}
func (a *appStoreMock) InitBuiltin(...*apps.App) {}

func (a *appStoreMock) Save(app *apps.App) error {
	a.apps[app.AppID] = app
	return nil
}
