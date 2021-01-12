// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

type Store interface {
	DeleteSub(*Subscription) error
	LoadApp(appID AppID) (*App, error)
	LoadSubs(subject Subject, teamID, channelID string) ([]*Subscription, error)
	ListApps() []*App
	StoreApp(app *App) error
	StoreSub(sub *Subscription) error
}
