// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"crypto/sha1" // nolint:gosec
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

type AppStore interface {
	config.Configurable

	AsMap() map[apps.AppID]*apps.App
	Delete(apps.AppID) error
	Get(appID apps.AppID) (*apps.App, error)
	InitBuiltin(...*apps.App)
	Save(app *apps.App) error
}

// appStore combines installed and builtin Apps.  The installed Apps are stored
// in KV store, and the list of their keys is stored in the config, as a map of
// AppID->sha1(App).
type appStore struct {
	*Service

	// mutex guards installed, the pointer to the map of locally-installed apps.
	mutex sync.RWMutex

	installed        map[apps.AppID]*apps.App
	builtinInstalled map[apps.AppID]*apps.App
}

var _ AppStore = (*appStore)(nil)

func (s *appStore) InitBuiltin(builtinApps ...*apps.App) {
	s.mutex.Lock()
	if s.builtinInstalled == nil {
		s.builtinInstalled = map[apps.AppID]*apps.App{}
	}
	for _, app := range builtinApps {
		s.builtinInstalled[app.AppID] = app
	}
	s.mutex.Unlock()
}

func (s *appStore) Configure(conf config.Config) {
	newInstalled := map[apps.AppID]*apps.App{}

	for id, key := range conf.InstalledApps {
		var app *apps.App
		err := s.mm.KV.Get(config.PrefixInstalledApp+key, &app)
		switch {
		case err != nil:
			s.mm.Log.Error(
				fmt.Sprintf("failed to load app %s: %s", id, err.Error()))

		case app == nil:
			s.mm.Log.Error(
				fmt.Sprintf("failed to load app %s: key %s not found", id, config.PrefixInstalledApp+key))

		default:
			newInstalled[apps.AppID(id)] = app
		}
	}

	s.mutex.Lock()
	s.installed = newInstalled
	s.mutex.Unlock()
}

func (s *appStore) Get(appID apps.AppID) (*apps.App, error) {
	s.mutex.RLock()
	installed := s.installed
	builtin := s.builtinInstalled
	s.mutex.RUnlock()

	app, ok := builtin[appID]
	if ok {
		return app, nil
	}
	app, ok = installed[appID]
	if ok {
		return app, nil
	}
	return nil, utils.ErrNotFound
}

func (s *appStore) AsMap() map[apps.AppID]*apps.App {
	s.mutex.RLock()
	installed := s.installed
	builtin := s.builtinInstalled
	s.mutex.RUnlock()

	out := map[apps.AppID]*apps.App{}
	for appID, app := range installed {
		out[appID] = app
	}
	for appID, app := range builtin {
		out[appID] = app
	}
	return out
}

func SortApps(appsMap map[apps.AppID]*apps.App) []*apps.App {
	out := []*apps.App{}
	for _, app := range appsMap {
		out = append(out, app)
	}

	sort.SliceStable(out, func(i, j int) bool {
		return apps.AppID(out[i].DisplayName) < apps.AppID(out[j].DisplayName)
	})
	return out
}

func (s *appStore) Save(app *apps.App) error {
	conf := s.conf.GetConfig()
	prevSHA := conf.InstalledApps[string(app.AppID)]

	data, err := json.Marshal(app)
	if err != nil {
		return err
	}
	sha := fmt.Sprintf("%x", sha1.Sum(data)) // nolint:gosec
	if sha == prevSHA {
		// no change in the data
		return nil
	}
	_, err = s.mm.KV.Set(config.PrefixInstalledApp+sha, app)
	if err != nil {
		return err
	}

	s.mutex.RLock()
	installed := s.installed
	s.mutex.RUnlock()
	updatedInstalled := map[apps.AppID]*apps.App{}
	for k, v := range installed {
		if k != app.AppID {
			updatedInstalled[k] = v
		}
	}
	updatedInstalled[app.AppID] = app
	s.mutex.Lock()
	s.installed = updatedInstalled
	s.mutex.Unlock()

	sc := conf.StoredConfig
	updated := map[string]string{}
	for k, v := range conf.InstalledApps {
		// delete prevSHA from the list by skipping
		if v != prevSHA {
			updated[k] = v
		}
	}
	updated[string(app.AppID)] = sha
	sc.InstalledApps = updated
	err = s.conf.StoreConfig(sc)
	if err != nil {
		return err
	}

	err = s.mm.KV.Delete(config.PrefixInstalledApp + prevSHA)
	if err != nil {
		s.mm.Log.Warn("failed to delete previous App KV value", "err", err.Error())
	}
	return nil
}

func (s *appStore) Delete(appID apps.AppID) error {
	s.mutex.RLock()
	installed := s.installed
	s.mutex.RUnlock()
	_, ok := installed[appID]
	if ok {
		return errors.Wrap(utils.ErrNotFound, string(appID))
	}

	conf := s.conf.GetConfig()
	sha, ok := conf.InstalledApps[string(appID)]
	if !ok {
		return utils.ErrNotFound
	}

	err := s.mm.KV.Delete(config.PrefixInstalledApp + sha)
	if err != nil {
		return err
	}

	updatedInstalled := map[apps.AppID]*apps.App{}
	for k, v := range installed {
		if k != appID {
			updatedInstalled[k] = v
		}
	}
	s.mutex.Lock()
	s.installed = updatedInstalled
	s.mutex.Unlock()

	sc := conf.StoredConfig
	updated := map[string]string{}
	for k, v := range conf.InstalledApps {
		updated[k] = v
	}
	delete(updated, string(appID))
	sc.InstalledApps = updated
	return s.conf.StoreConfig(sc)
}
