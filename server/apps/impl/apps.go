// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package impl

import (
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

func (s *service) ListApps() []*apps.App {
	conf := s.Configurator.GetConfig()
	out := []*apps.App{}
	if len(conf.Apps) == 0 {
		return out
	}
	for _, v := range conf.Apps {
		app := apps.AppFromConfigMap(v)
		out = append(out, app)
	}
	return out
}

func (s *service) GetApp(appID apps.AppID) (*apps.App, error) {
	conf := s.Configurator.GetConfig()
	if len(conf.Apps) == 0 {
		return nil, utils.ErrNotFound
	}
	v := conf.Apps[string(appID)]
	if v == nil {
		return nil, utils.ErrNotFound
	}
	return apps.AppFromConfigMap(v), nil
}

func (s *service) StoreApp(app *apps.App) error {
	conf := s.Configurator.GetConfig()
	if len(conf.Apps) == 0 {
		conf.Apps = map[string]interface{}{}
	}

	conf.Apps[string(app.Manifest.AppID)] = app.ConfigMap()

	// Refresh the local config immediately, do not wait for the
	// OnConfigurationChange.
	err := s.Configurator.RefreshConfig(conf.StoredConfig)
	if err != nil {
		return err
	}

	return s.Configurator.StoreConfig(conf.StoredConfig)
}
