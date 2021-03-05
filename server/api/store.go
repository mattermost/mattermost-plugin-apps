// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/awsclient"
)

type Store interface {
	App() AppStore
	Sub() SubStore
	Manifest() ManifestStore
}

type AppStore interface {
	Configurable

	AsMap() map[apps.AppID]*apps.App
	Delete(apps.AppID) error
	Get(appID apps.AppID) (*apps.App, error)
	InitBuiltin(...*apps.App)
	Save(app *apps.App) error
}

type ManifestStore interface {
	Configurable

	AsMap() map[apps.AppID]*apps.Manifest
	DeleteLocal(apps.AppID) error
	Get(apps.AppID) (*apps.Manifest, error)
	InitGlobal(_ awsclient.Client, bucket string) error
	StoreLocal(*apps.Manifest) error
}
type SubStore interface {
	Get(subject apps.Subject, teamID, channelID string) ([]*apps.Subscription, error)
	Save(sub *apps.Subscription) error
	Delete(*apps.Subscription) error
}
