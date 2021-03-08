// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import "github.com/mattermost/mattermost-plugin-apps/apps"

type Store interface {
	App() AppStore
	Sub() SubStore
	Manifest() ManifestStore
}

type AppStore interface {
	Get(appID apps.AppID) (*apps.App, error)
	GetAll() []*apps.App
	Save(app *apps.App) error
	Delete(app *apps.App) error
}

type SubStore interface {
	Get(subject apps.Subject, teamID, channelID string) ([]*apps.Subscription, error)
	Save(sub *apps.Subscription) error
	Delete(*apps.Subscription) error
}

type ManifestStore interface {
	Get(appID apps.AppID) (*apps.Manifest, error)
	GetAll() map[apps.AppID]*apps.Manifest
	Save(manifest *apps.Manifest)
	Delete(appID apps.AppID)
	Cleanup()
}
