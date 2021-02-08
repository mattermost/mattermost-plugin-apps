// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

const prefixSubs = "sub_"

type Store struct {
	mm     *pluginapi.Client
	conf   api.Configurator
	stores Stores
}

type Stores struct {
	app      api.AppStore
	sub      api.SubStore
	manifest api.ManifestStore
}

var _ api.Store = (*Store)(nil)

func New(mm *pluginapi.Client, conf api.Configurator) *Store {
	store := &Store{
		mm:   mm,
		conf: conf,
	}
	store.stores.app = newAppStore(store)
	store.stores.sub = newSubStore(store)
	store.stores.manifest = newManifestStore()
	return store
}

func (s *Store) App() api.AppStore {
	return s.stores.app
}

func (s *Store) Sub() api.SubStore {
	return s.stores.sub
}

func (s *Store) Manifest() api.ManifestStore {
	return s.stores.manifest
}
