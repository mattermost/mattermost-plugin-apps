// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"sort"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-server/v6/plugin"
)

type FilterOpt bool

const (
	EnabledAppsOnly = FilterOpt(false)
	AllApps         = FilterOpt(true)
)

// appStore combines installed and builtin Apps.  The installed Apps are stored
// in KV store, and the list of their keys is stored in the config, as a map of
// AppID->sha1(App).
type AppStore struct {
	schemaVersion string
	builtin       map[apps.AppID]apps.App
	installed     *CachedStore[apps.App]
}

type Apps interface {
	AsList(filter FilterOpt) []apps.App
	AsMap(filter FilterOpt) map[apps.AppID]apps.App
	Delete(r *incoming.Request, appID apps.AppID) error
	Get(appID apps.AppID) (*apps.App, error)
	Save(r *incoming.Request, app apps.App) error
}

var _ Apps = (*AppStore)(nil)

func MakeAppStore(api plugin.API, conf config.Service, builtinApps ...apps.App) (*AppStore, error) {
	store, err := MakeCachedStore[apps.App](AppStoreName, api, conf)
	if err != nil {
		return nil, err
	}
	builtin := map[apps.AppID]apps.App{}
	for _, app := range builtinApps {
		app.DeployType = apps.DeployBuiltin
		builtin[app.AppID] = app
	}
	return &AppStore{
		schemaVersion: conf.Get().PluginManifest.Version,
		builtin:       builtin,
		installed:     store,
	}, nil
}

func (s *AppStore) Get(appID apps.AppID) (*apps.App, error) {
	app, ok := s.builtin[appID]
	if ok {
		return &app, nil
	}
	app, ok = s.installed.Get(string(appID))
	if ok {
		return &app, nil
	}
	return nil, utils.NewNotFoundError("app %s is not installed", appID)
}

func (s *AppStore) AsMap(filter FilterOpt) map[apps.AppID]apps.App {
	out := map[apps.AppID]apps.App{}
	for id := range s.installed.Index() {
		if app, ok := s.installed.Get(id); ok {
			if filter == AllApps || !app.Disabled {
				out[apps.AppID(id)] = app
			}
		}
	}
	for appID, app := range s.builtin {
		if filter == AllApps || !app.Disabled {
			out[appID] = app
		}
	}
	return out
}

func (s *AppStore) AsList(filter FilterOpt) []apps.App {
	var out []apps.App
	for _, app := range s.AsMap(filter) {
		out = append(out, app)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return apps.AppID(out[i].DisplayName) < apps.AppID(out[j].DisplayName)
	})
	return out
}

func (s *AppStore) Save(r *incoming.Request, app apps.App) error {
	app.Manifest.SchemaVersion = s.schemaVersion
	return s.installed.Put(r, string(app.AppID), app)
}

func (s *AppStore) Delete(r *incoming.Request, appID apps.AppID) error {
	return s.installed.Delete(r, string(appID))
}
