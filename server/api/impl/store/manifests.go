// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

func (s *Store) EmptyManifests() {
	s.manifests = map[api.AppID]*api.Manifest{}
}

func (s *Store) StoreManifest(manifest *api.Manifest) {
	s.manifests[manifest.AppID] = manifest
}

func (s *Store) LoadManifest(appID api.AppID) (*api.Manifest, error) {
	manifest, ok := s.manifests[appID]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return manifest, nil
}

func (s *Store) ListManifests() map[api.AppID]*api.Manifest {
	return s.manifests
}

func (s *Store) populateAppWithManifest(app *api.App) *api.App {
	manifest, ok := s.manifests[app.ID]
	if !ok {
		s.mm.Log.Error("This should not have happened. No manifest avaliable for", "app_id", app.ID)
	}
	app.Manifest = manifest
	return app
}
