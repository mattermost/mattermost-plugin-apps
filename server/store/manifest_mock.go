package store

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/aws"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

type manifestStoreMock struct {
	manifests map[apps.AppID]*apps.Manifest
}

var _ ManifestStore = (*manifestStoreMock)(nil)

func NewManifestStoreMock(manifests map[apps.AppID]*apps.Manifest) ManifestStore {
	return &manifestStoreMock{manifests: manifests}
}

func (m *manifestStoreMock) AsMap() map[apps.AppID]*apps.Manifest {
	return m.manifests
}

func (m *manifestStoreMock) Configure(c config.Config) {}

func (m *manifestStoreMock) DeleteLocal(appID apps.AppID) error {
	if _, ok := m.manifests[appID]; !ok {
		return utils.ErrNotFound
	}
	delete(m.manifests, appID)
	return nil
}

func (m *manifestStoreMock) Get(appID apps.AppID) (*apps.Manifest, error) {
	manifest, ok := m.manifests[appID]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return manifest, nil
}

func (m *manifestStoreMock) InitGlobal(_ aws.Client, bucket string) error {
	return nil
}

func (m *manifestStoreMock) StoreLocal(manifest *apps.Manifest) error {
	m.manifests[manifest.AppID] = manifest
	return nil
}
