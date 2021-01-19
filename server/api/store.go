// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import "github.com/mattermost/mattermost-plugin-apps/modelapps"

type Store interface {
	DeleteSub(*modelapps.Subscription) error
	LoadApp(appID modelapps.AppID) (*modelapps.App, error)
	LoadSubs(subject modelapps.Subject, teamID, channelID string) ([]*modelapps.Subscription, error)
	ListApps() []*modelapps.App
	StoreApp(app *modelapps.App) error
	StoreSub(sub *modelapps.Subscription) error
}
