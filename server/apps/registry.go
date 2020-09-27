// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

type Registry interface {
	Store(*App) error
	Get(AppID) (*App, error)
}

type registry struct {
	configurator configurator.Service

	// <><> Needs to come from config to be synchronized, or read from KV every request, sync.Map is unnecessary
	apps map[AppID]*App
}

var _ Registry = (*registry)(nil)

func NewRegistry(configurator configurator.Service) Registry {
	return &registry{
		configurator: configurator,
		apps:         map[AppID]*App{},
	}
}

func (r *registry) Store(app *App) error {
	// <><> TODO remove mock, implement for real.
	r.apps[app.Manifest.AppID] = app
	return nil
}

func (r *registry) Get(appID AppID) (*App, error) {
	app, found := r.apps[appID]
	if !found {
		return nil, utils.ErrNotFound
	}
	return app, nil
}
