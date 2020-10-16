// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
)

const prefixApp = "app_"
const prefixSubs = "sub_"

type Service interface {
	DeleteApp(AppID) error
	DeleteSub(*Subscription) error
	GetApp(AppID) (*App, error)
	GetAllAppIDs() ([]AppID, error)
	GetSubs(subject Subject, teamID, channelID string) ([]*Subscription, error)
	StoreApp(*App) error
	StoreSub(sub *Subscription) error
}

type store struct {
	Mattermost   *pluginapi.Client
	Configurator configurator.Service
}

func NewService(mm *pluginapi.Client, conf configurator.Service) Service {
	return &store{
		Mattermost:   mm,
		Configurator: conf,
	}
}
