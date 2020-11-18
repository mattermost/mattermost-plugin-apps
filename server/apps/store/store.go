// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
)

const prefixSubs = "sub_"

type Service interface {
	DeleteSub(*apps.Subscription) error
	GetSubs(subject apps.Subject, teamID, channelID string) ([]*apps.Subscription, error)
	StoreSub(sub *apps.Subscription) error
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
