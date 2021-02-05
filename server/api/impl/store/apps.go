// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

func (s *Store) ListApps() []*api.App {
	conf := s.conf.GetConfig()
	out := []*api.App{}
	if len(conf.Apps) == 0 {
		return out
	}
	for _, v := range conf.Apps {
		app := api.AppFromConfigMap(v)
		app = s.populateAppWithManifest(app)
		out = append(out, app)
	}
	return out
}

func (s *Store) LoadApp(appID api.AppID) (*api.App, error) {
	conf := s.conf.GetConfig()
	if len(conf.Apps) == 0 {
		return nil, utils.ErrNotFound
	}
	v := conf.Apps[string(appID)]
	if v == nil {
		return nil, utils.ErrNotFound
	}
	app := api.AppFromConfigMap(v)
	app = s.populateAppWithManifest(app)
	return app, nil
}

func (s *Store) StoreApp(app *api.App) error {
	conf := s.conf.GetConfig()
	if len(conf.Apps) == 0 {
		conf.Apps = map[string]interface{}{}
	}
	//do not store manifest in the config
	app.ID = app.Manifest.AppID
	app.Manifest = nil

	conf.Apps[string(app.Manifest.AppID)] = app.ConfigMap()

	// Refresh the local config immediately, do not wait for the
	// OnConfigurationChange.
	err := s.conf.RefreshConfig(conf.StoredConfig)
	if err != nil {
		return err
	}

	return s.conf.StoreConfig(conf.StoredConfig)
}

func (s *Store) DeleteApp(app *api.App) error {
	conf := s.conf.GetConfig()
	delete(conf.Apps, string(app.Manifest.AppID))

	// Refresh the local config immediately, do not wait for the
	// OnConfigurationChange.
	err := s.conf.RefreshConfig(conf.StoredConfig)
	if err != nil {
		return err
	}
	return s.conf.StoreConfig(conf.StoredConfig)
}
