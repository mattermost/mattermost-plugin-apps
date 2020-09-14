// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package cloudapps

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-cloudapps/server/configurator"
)

type Service struct {
	Config        configurator.Service
	Expander      Expander
	Mattermost    *pluginapi.Client
	Proxy         Proxy
	Registry      Registry
	Subscriptions Subscriptions
}

func NewService(mm *pluginapi.Client, configurator configurator.Service) *Service {
	registry := NewRegistry(configurator)
	expander := NewExpander(mm, configurator)
	subs := NewSubscriptions(configurator)
	proxy := NewProxy(mm, configurator, subs)

	return &Service{
		Config:        configurator,
		Expander:      expander,
		Mattermost:    mm,
		Proxy:         proxy,
		Registry:      registry,
		Subscriptions: subs,
	}
}
