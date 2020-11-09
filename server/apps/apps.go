// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

func (s *service) ListApps() []*api.App {
	conf := s.Configurator.GetConfig()
	out := []*api.App{}
	for _, v := range conf.Apps {
		app := api.AppFromConfigMap(v)
		out = append(out, app)
	}
	return out
}

func (s *service) GetApp(appID api.AppID) (*api.App, error) {
	conf := s.Configurator.GetConfig()
	if len(conf.Apps) == 0 {
		return nil, utils.ErrNotFound
	}
	v := conf.Apps[string(appID)]
	if v == nil {
		return nil, utils.ErrNotFound
	}
	return api.AppFromConfigMap(v), nil
}

func (s *service) StoreApp(app *api.App) error {
	conf := s.Configurator.GetConfig()
	if len(conf.Apps) == 0 {
		conf.Apps = map[string]interface{}{}
	}

	conf.Apps[string(app.Manifest.AppID)] = app.ConfigMap()

	// Refresh the local config immediately, do not wait for the
	// OnConfigurationChange.
	err := s.Configurator.Refresh(conf.StoredConfig)
	if err != nil {
		return err
	}

	return s.Configurator.Store(conf.StoredConfig)
}
