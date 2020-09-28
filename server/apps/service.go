// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
)

type Service struct {
	Configurator  configurator.Service
	Mattermost    *pluginapi.Client
	Expander      Expander
	Registry      Registry
	Subscriptions Subscriptions
	Client        Client
	Hooks         Hooks
	API           API
}

func NewService(mm *pluginapi.Client, configurator configurator.Service) *Service {
	registry := NewRegistry(configurator)
	expander := NewExpander(mm, configurator)
	subs := NewSubscriptions(configurator)

	s := &Service{
		Configurator:  configurator,
		Mattermost:    mm,
		Expander:      expander,
		Registry:      registry,
		Subscriptions: subs,
	}
	s.Hooks = s
	s.Client = s
	s.API = s

	return s
}
