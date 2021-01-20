// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import "github.com/mattermost/mattermost-plugin-apps/apps"

type Store interface {
	DeleteSub(*apps.Subscription) error
	LoadApp(appID apps.AppID) (*apps.App, error)
	LoadSubs(subject apps.Subject, teamID, channelID string) ([]*apps.Subscription, error)
	ListApps() []*apps.App
	StoreApp(app *apps.App) error
	StoreSub(sub *apps.Subscription) error
}
