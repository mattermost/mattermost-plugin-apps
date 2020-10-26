// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

type SessionToken string

type Service struct {
	Configurator configurator.Service
	Mattermost   *pluginapi.Client
	Store        store.Service
	API          api.API
	Client       Client
}

type service struct {
	Service
}

func NewService(mm *pluginapi.Client, configurator configurator.Service) *Service {
	s := &service{
		Service: Service{
			Store:        store.NewService(mm, configurator),
			Configurator: configurator,
			Mattermost:   mm,
		},
	}
	s.Client = newClient(s.Store)
	s.API = s

	return &s.Service
}
