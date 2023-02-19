// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"sort"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
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
	CachedStore[apps.App]
}

type Apps interface {
	AsList(FilterOpt) []apps.App
	AsMap(FilterOpt) map[apps.AppID]apps.App
	Delete(*incoming.Request, apps.AppID) error
	Get(apps.AppID) (*apps.App, error)
	InitBuiltin(builtinApps ...apps.App)
	Save(*incoming.Request, apps.App) error
}

var _ Apps = (*AppStore)(nil)

func (s *Service) makeAppStore(version string, log utils.Logger) (*AppStore, error) {
	store, err := MakeCachedStore[apps.App](AppStoreName, s.cluster, log)
	if err != nil {
		return nil, err
	}
	return &AppStore{
		schemaVersion: version,
		CachedStore:   store,
	}, nil
}

func (s *AppStore) InitBuiltin(builtinApps ...apps.App) {
	builtin := map[apps.AppID]apps.App{}
	for _, app := range builtinApps {
		app.DeployType = apps.DeployBuiltin
		builtin[app.AppID] = app
	}
	s.builtin = builtin
}

func (s *AppStore) Get(appID apps.AppID) (*apps.App, error) {
	if app, ok := s.builtin[appID]; ok {
		return &app, nil
	}
	if app := s.CachedStore.Get(string(appID)); app != nil {
		return app, nil
	}
	return nil, utils.NewNotFoundError("app %s is not installed", appID)
}

func (s *AppStore) AsMap(filter FilterOpt) map[apps.AppID]apps.App {
	out := map[apps.AppID]apps.App{}
	for id := range s.Index() {
		if app := s.CachedStore.Get(id); app != nil {
			if filter == AllApps || !app.Disabled {
				out[apps.AppID(id)] = *app
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
	return s.CachedStore.Put(r, string(app.AppID), &app)
}

func (s *AppStore) Delete(r *incoming.Request, appID apps.AppID) error {
	return s.CachedStore.Put(r, string(appID), nil)
}
