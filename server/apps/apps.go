// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"errors"

	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

func (s *Service) ListApps() []*App {
	conf := s.Configurator.GetConfig()
	out := []*App{}
	if len(conf.Apps) == 0 {
		return out
	}
	for _, v := range conf.Apps {
		app := AppFromConfigMap(v)
		out = append(out, app)
	}
	return out
}

func (s *Service) GetApp(appID AppID) (*App, error) {
	conf := s.Configurator.GetConfig()
	if len(conf.Apps) == 0 {
		return nil, utils.ErrNotFound
	}
	v := conf.Apps[string(appID)]
	if v == nil {
		return nil, utils.ErrNotFound
	}
	return AppFromConfigMap(v), nil
}

func (s *Service) StoreApp(app *App) error {
	conf := s.Configurator.GetConfig()
	if conf.StoredConfig == nil {
		return errors.New("conf is nil")
	}

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
