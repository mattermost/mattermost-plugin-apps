// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

func (s *store) StoreApp(app *App) error {
	_, err := s.Mattermost.KV.Set(prefixApp+string(app.Manifest.AppID), app)
	return err
}

func (s *store) GetApp(appID AppID) (*App, error) {
	var app *App
	err := s.Mattermost.KV.Get(prefixApp+string(appID), &app)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, utils.ErrNotFound
	}
	return app, nil
}

func (s *store) DeleteApp(appID AppID) error {
	return s.Mattermost.KV.Delete(prefixApp + string(appID))
}

// TODO SLOW 0/5 put the list of installed Apps in the (Mattermost) Config
func (s *store) ListApps() ([]AppID, error) {
	appIDs := []AppID{}
	for i := 0; ; i++ {
		keys, err := s.Mattermost.KV.ListKeys(i, 1000)
		if err != nil {
			return nil, err
		}
		if len(keys) == 0 {
			break
		}

		for _, key := range keys {
			if strings.HasPrefix(key, prefixApp) {
				appIDs = append(appIDs, AppID(key[len(prefixApp):]))
			}
		}
	}
	return appIDs, nil
}
