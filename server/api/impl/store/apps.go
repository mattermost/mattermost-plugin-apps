// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

type AppStore struct {
	*Store
}

var _ api.AppStore = (*AppStore)(nil)

func newAppStore(st *Store) api.AppStore {
	s := &AppStore{st}
	return s
}

func (s AppStore) GetAll() []*apps.App {
	conf := s.conf.GetConfig()
	out := []*apps.App{}
	if len(conf.Apps) == 0 {
		return out
	}
	for _, v := range conf.Apps {
		app := apps.AppFromConfigMap(v)
		app = s.populateAppWithManifest(app)
		out = append(out, app)
	}
	return out
}

func (s AppStore) Get(appID apps.AppID) (*apps.App, error) {
	conf := s.conf.GetConfig()
	if len(conf.Apps) == 0 {
		return nil, utils.ErrNotFound
	}
	v := conf.Apps[string(appID)]
	if v == nil {
		return nil, utils.ErrNotFound
	}
	app := apps.AppFromConfigMap(v)
	app = s.populateAppWithManifest(app)
	return app, nil
}

func (s AppStore) Save(app *apps.App) error {
	conf := s.conf.GetConfig()
	if len(conf.Apps) == 0 {
		conf.Apps = map[string]interface{}{}
	}

	// Copy app before modifying it
	cApp := &apps.App{}
	*cApp = *app
	// do not store manifest in the config
	cApp.AppID = app.Manifest.AppID
	cApp.Manifest = nil

	conf.Apps[string(app.AppID)] = app.ConfigMap()

	// Refresh the local config immediately, do not wait for the
	// OnConfigurationChange.
	err := s.conf.RefreshConfig(conf.StoredConfig)
	if err != nil {
		return err
	}

	return s.conf.StoreConfig(conf.StoredConfig)
}

func (s AppStore) Delete(app *apps.App) error {
	conf := s.conf.GetConfig()
	delete(conf.Apps, string(app.AppID))
	s.stores.manifest.Delete(app.AppID)

	// Refresh the local config immediately, do not wait for the
	// OnConfigurationChange.
	err := s.conf.RefreshConfig(conf.StoredConfig)
	if err != nil {
		return err
	}
	return s.conf.StoreConfig(conf.StoredConfig)
}

func (s AppStore) populateAppWithManifest(app *apps.App) *apps.App {
	manifest, err := s.stores.manifest.Get(app.AppID)
	if err != nil {
		s.mm.Log.Error("This should not have happened. No manifest available for", "app_id", app.AppID)
	}
	app.Manifest = manifest
	return app
}
