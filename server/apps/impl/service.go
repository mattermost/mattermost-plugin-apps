// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package impl

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/apps/store"
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
)

type service struct {
	apps.Service
	Store store.Service
}

func NewService(mm *pluginapi.Client, configurator configurator.Service) *apps.Service {
	s := &service{
		Service: apps.Service{
			Configurator: configurator,
			Mattermost:   mm,
		},
		Store: store.NewService(mm, configurator),
	}
	s.Client = newClient(s.Store)
	s.API = s

	return &s.Service
}
