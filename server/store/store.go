// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
)

const prefixApp = "app_"
const prefixSubs = "sub_"

type Service interface {
	DeleteApp(api.AppID) error
	DeleteSub(*api.Subscription) error
	GetApp(api.AppID) (*api.App, error)
	ListApps() ([]api.AppID, error)
	GetSubs(subject api.Subject, teamID, channelID string) ([]*api.Subscription, error)
	StoreApp(*api.App) error
	StoreSub(sub *api.Subscription) error
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
