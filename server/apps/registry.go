// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/pkg/errors"
)

type Registry interface {
	Store(*App) error
	Get(AppID) (*App, error)
	GetLocations(appID AppID) ([]*LocationRegistry, error)
	GetAllLocations() ([]*LocationRegistry, error)
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

func (r *registry) GetLocations(appID AppID) ([]*LocationRegistry, error) {
	app, err := r.Get(appID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get app")
	}

	return app.Manifest.Locations, nil
}

func (r *registry) GetAllLocations() ([]*LocationRegistry, error) {
	locations := []*LocationRegistry{}
	for appID := range r.apps {
		loc, err := r.GetLocations(appID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get single app location registers")
		}
		locations = append(locations, loc...)
	}
	return locations, nil
}
