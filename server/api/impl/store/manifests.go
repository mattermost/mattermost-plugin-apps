// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

func (s *Store) EmptyManifests() {
	s.manifests = map[apps.AppID]*apps.Manifest{}
}

func (s *Store) StoreManifest(manifest *apps.Manifest) {
	s.manifests[manifest.AppID] = manifest
}

func (s *Store) LoadManifest(appID apps.AppID) (*apps.Manifest, error) {
	manifest, ok := s.manifests[appID]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return manifest, nil
}

func (s *Store) ListManifests() map[apps.AppID]*apps.Manifest {
	return s.manifests
}

func (s *Store) populateAppWithManifest(app *apps.App) *apps.App {
	manifest, ok := s.manifests[app.ID]
	if !ok {
		s.mm.Log.Error("This should not have happened. No manifest avaliable for", "app_id", app.ID)
	}
	app.Manifest = manifest
	return app
}
