// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"sort"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type TestAppStore map[apps.AppID]apps.App

var _ AppStore = (TestAppStore)(nil)

func (s TestAppStore) InitBuiltin(builtinApps ...apps.App) {
	for _, app := range builtinApps {
		app.DeployType = apps.DeployBuiltin
		s[app.AppID] = app
	}
}

func (s TestAppStore) Get(appID apps.AppID) (*apps.App, error) {
	app, ok := s[appID]
	if ok {
		return &app, nil
	}
	return nil, utils.NewNotFoundError("app %s is not installed", appID)
}

func (s TestAppStore) AsMap(filter FilterOpt) map[apps.AppID]apps.App {
	out := map[apps.AppID]apps.App{}
	for id, app := range s {
		if filter == AllApps || !app.Disabled {
			out[id] = app
		}
	}
	return out
}

func (s TestAppStore) AsList(filter FilterOpt) []apps.App {
	var out []apps.App
	for _, app := range s.AsMap(filter) {
		out = append(out, app)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return apps.AppID(out[i].DisplayName) < apps.AppID(out[j].DisplayName)
	})
	return out
}

func (s TestAppStore) Save(r *incoming.Request, app apps.App) error {
	s[app.AppID] = app
	return nil
}

func (s TestAppStore) Delete(r *incoming.Request, appID apps.AppID) error {
	delete(s, appID)
	return nil
}
