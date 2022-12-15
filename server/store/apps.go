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
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type AppStore interface {
	config.Configurable

	InitBuiltin(...apps.App)

	Get(apps.AppID) (*apps.App, error)
	List() []apps.App                   // List returns all installed, enabled apps.
	ListAsMap() map[apps.AppID]apps.App // List returns all installed, enabled apps as a map.
	ListIncludeDisabled() []apps.App    // List returns all installed apps, including the disabled ones.
	Save(*incoming.Request, apps.App) error
	Delete(*incoming.Request, apps.AppID) error
}

// appStore combines installed and builtin Apps.  The installed Apps are stored
// in KV store, and the list of their keys is stored in the config, as a map of
// AppID->sha1(App).
type appStore struct {
	*Service

	// mutex guards installed, the pointer to the map of locally-installed apps.
	mutex sync.RWMutex

	installed        map[apps.AppID]apps.App
	builtinInstalled map[apps.AppID]apps.App
}

var _ AppStore = (*appStore)(nil)

func (s *Service) makeAppStore(conf config.Config) (*appStore, error) {
	appStore := &appStore{Service: s}
	err := appStore.Configure(conf, s.conf.NewBaseLogger())
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize App store")
	}
	return appStore, nil
}

func (s *appStore) InitBuiltin(builtinApps ...apps.App) {
	s.mutex.Lock()
	if s.builtinInstalled == nil {
		s.builtinInstalled = map[apps.AppID]apps.App{}
	}
	for _, app := range builtinApps {
		app.DeployType = apps.DeployBuiltin
		s.builtinInstalled[app.AppID] = app
	}
	s.mutex.Unlock()
}

func (s *appStore) Configure(conf config.Config, log utils.Logger) error {
	newInstalled := map[apps.AppID]apps.App{}
	mm := s.conf.MattermostAPI()

	for id, key := range conf.InstalledApps {
		log = log.With("app_id", id)

		var data []byte
		err := mm.KV.Get(KVInstalledAppPrefix+key, &data)
		if err != nil {
			log.WithError(err).Errorw("failed to load app")
			continue
		}

		if len(data) == 0 {
			err = utils.NewNotFoundError(KVInstalledAppPrefix + key)
			log.WithError(err).Errorw("failed to load app")
			continue
		}

		app, err := apps.DecodeCompatibleApp(data)
		if err != nil {
			log.WithError(err).Errorw("failed to decode app")
			continue
		}
		newInstalled[apps.AppID(id)] = *app
	}

	s.mutex.Lock()
	s.installed = newInstalled
	s.mutex.Unlock()
	return nil
}

func (s *appStore) Get(appID apps.AppID) (*apps.App, error) {
	s.mutex.RLock()
	installed := s.installed
	builtin := s.builtinInstalled
	s.mutex.RUnlock()

	app, ok := builtin[appID]
	if ok {
		return &app, nil
	}
	app, ok = installed[appID]
	if ok {
		return &app, nil
	}
	return nil, utils.NewNotFoundError("app %s is not installed", appID)
}

func (s *appStore) List() []apps.App {
	var out []apps.App
	for _, app := range s.ListAsMap() {
		out = append(out, app)
	}

	return out
}

func (s *appStore) ListAsMap() map[apps.AppID]apps.App {
	s.mutex.RLock()
	installed := s.installed
	builtin := s.builtinInstalled
	s.mutex.RUnlock()

	out := map[apps.AppID]apps.App{}
	for appID, app := range installed {
		if !app.Disabled {
			out[appID] = app
		}
	}
	for appID, app := range builtin {
		if !app.Disabled {
			out[appID] = app
		}
	}

	return out
}

func (s *appStore) ListIncludeDisabled() []apps.App {
	s.mutex.RLock()
	installed := s.installed
	builtin := s.builtinInstalled
	s.mutex.RUnlock()

	out := []apps.App{}
	for _, app := range installed {
		out = append(out, app)
	}
	for _, app := range builtin {
		out = append(out, app)
	}

	return out
}

// SortApps sorts a list of apps alphabetically, by display name.
func SortApps(appList []apps.App) []apps.App {
	sort.SliceStable(appList, func(i, j int) bool {
		return apps.AppID(appList[i].DisplayName) < apps.AppID(appList[j].DisplayName)
	})

	return appList
}

func (s *appStore) Save(r *incoming.Request, app apps.App) error {
	conf := s.conf.Get()
	mm := s.conf.MattermostAPI()
	prevSHA := conf.InstalledApps[string(app.AppID)]

	app.Manifest.SchemaVersion = conf.PluginManifest.Version
	data, err := json.Marshal(app)
	if err != nil {
		return err
	}
	sha := fmt.Sprintf("%x", sha1.Sum(data)) // nolint:gosec
	_, err = mm.KV.Set(KVInstalledAppPrefix+sha, app)
	if err != nil {
		return err
	}

	s.mutex.RLock()
	installed := s.installed
	s.mutex.RUnlock()
	updatedInstalled := map[apps.AppID]apps.App{}
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
	err = s.conf.StoreConfig(sc, r.Log)
	if err != nil {
		return err
	}

	if sha != prevSHA {
		err = mm.KV.Delete(KVInstalledAppPrefix + prevSHA)
		if err != nil {
			r.Log.WithError(err).Warnf("Failed to delete previous App KV value")
		}
	}

	return nil
}

func (s *appStore) Delete(r *incoming.Request, appID apps.AppID) error {
	s.mutex.RLock()
	installed := s.installed
	s.mutex.RUnlock()
	_, ok := installed[appID]
	if !ok {
		return utils.NewNotFoundError(appID)
	}

	conf := s.conf.Get()
	mm := s.conf.MattermostAPI()
	sha, ok := conf.InstalledApps[string(appID)]
	if !ok {
		return utils.ErrNotFound
	}

	err := mm.KV.Delete(KVInstalledAppPrefix + sha)
	if err != nil {
		return err
	}

	updatedInstalled := map[apps.AppID]apps.App{}
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
	return s.conf.StoreConfig(sc, r.Log)
}
