// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package impl

import (
	"sync"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/apps/store"
	"github.com/mattermost/mattermost-plugin-apps/server/aws"
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
)

type service struct {
	apps.Service
	Store store.Service

	appsCache *sync.Map
}

func NewService(mm *pluginapi.Client, configurator configurator.Service) *apps.Service {
	s := &service{
		Service: apps.Service{
			Configurator: configurator,
			Mattermost:   mm,
		},
		appsCache: &sync.Map{},
		Store:     store.NewService(mm, configurator),
	}
	s.Client = s.newClient()
	s.API = s
	s.AWSProxy = aws.NewAWSProxy(mm)

	return &s.Service
}
