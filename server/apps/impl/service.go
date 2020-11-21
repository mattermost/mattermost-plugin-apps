// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package impl

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/apps/store"
)

type service struct {
	apps.Service
	store apps.Store
}

func NewService(mm *pluginapi.Client, conf apps.Configurator) *apps.Service {
	s := &service{
		Service: apps.Service{
			Mattermost:   mm,
			Configurator: conf,
		},
		store: store.NewStore(mm, conf),
	}
	s.Service.API = s
	s.Service.Upstream = s
	return &s.Service
}
