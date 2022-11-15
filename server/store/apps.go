// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"sort"

	"github.com/mattermost/mattermost-server/v6/plugin"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type AppStore interface {
	InitBuiltin(...apps.App)

	Get(apps.AppID) (*apps.App, error)
	AsMap() map[apps.AppID]apps.App
	Save(*incoming.Request, apps.App) error
	Delete(*incoming.Request, apps.AppID) error
}

// appStore combines installed and builtin Apps.  The installed Apps are stored
// in KV store, and the list of their keys is stored in the config, as a map of
// AppID->sha1(App).
type appStore struct {
	schemaVersion string
	cache         *CachedStore[apps.App]

	builtins map[apps.AppID]apps.App
}

var _ AppStore = (*appStore)(nil)

func makeAppStore(conf config.Service, api plugin.API, logger utils.Logger) (*appStore, error) {
	s, err := MakeCachedStore[apps.App](AppStoreName, api, conf.MattermostAPI(), logger)
	if err != nil {
		return nil, err
	}
	return &appStore{
		cache:         s,
		schemaVersion: conf.Get().PluginManifest.Version,
	}, nil
}

func (s *appStore) InitBuiltin(builtinApps ...apps.App) {
	if s.builtins == nil {
		s.builtins = map[apps.AppID]apps.App{}
	}
	for _, app := range builtinApps {
		app.DeployType = apps.DeployBuiltin
		s.builtins[app.AppID] = app
	}
}

func (s *appStore) Get(appID apps.AppID) (*apps.App, error) {
	app, ok := s.builtins[appID]
	if ok {
		return &app, nil
	}
	app, ok = s.cache.Get(string(appID))
	if ok {
		return &app, nil
	}
	return nil, utils.NewNotFoundError("app %s is not installed", appID)
}

func (s *appStore) AsMap() map[apps.AppID]apps.App {
	out := map[apps.AppID]apps.App{}
	for id := range s.cache.Index() {
		if app, ok := s.cache.Get(id); ok {
			out[apps.AppID(id)] = app
		}
	}
	for appID, app := range s.builtins {
		out[appID] = app
	}
	return out
}

func SortApps(appsMap map[apps.AppID]apps.App) []apps.App {
	out := []apps.App{}
	for _, app := range appsMap {
		out = append(out, app)
	}

	sort.SliceStable(out, func(i, j int) bool {
		return apps.AppID(out[i].DisplayName) < apps.AppID(out[j].DisplayName)
	})
	return out
}

func (s *appStore) Save(r *incoming.Request, app apps.App) error {
	app.Manifest.SchemaVersion = s.schemaVersion
	return s.cache.Put(r, string(app.AppID), app)
}

func (s *appStore) Delete(r *incoming.Request, appID apps.AppID) error {
	return s.cache.Delete(r, string(appID))
}
