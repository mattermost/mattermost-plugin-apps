// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

type ManifestStore struct {
	manifests map[apps.AppID]*apps.Manifest
}

var _ api.ManifestStore = (*ManifestStore)(nil)

func newManifestStore() api.ManifestStore {
	manifests := map[apps.AppID]*apps.Manifest{}
	s := &ManifestStore{manifests}
	return s
}

func (s ManifestStore) Cleanup() {
	s.manifests = map[apps.AppID]*apps.Manifest{}
}

func (s ManifestStore) Save(manifest *apps.Manifest) {
	s.manifests[manifest.AppID] = manifest
}

func (s ManifestStore) Get(appID apps.AppID) (*apps.Manifest, error) {
	manifest, ok := s.manifests[appID]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return manifest, nil
}

func (s ManifestStore) GetAll() map[apps.AppID]*apps.Manifest {
	return s.manifests
}
