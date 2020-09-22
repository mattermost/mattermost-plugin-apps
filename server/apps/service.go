// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"net/http"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
)

type Service interface {
	OnUserJoinedChannel(pluginContext *plugin.Context, channelMember *model.ChannelMember, actor *model.User)
}

type Proxy interface {
	SendChangeNotification(s *Subscription, msg interface{})
	Post(app *App, url string, msg interface{}) (*http.Response, error)
}

type Subscriptions interface {
	GetSubscriptionsForChannel(subj SubscriptionSubject, channelID string) ([]*Subscription, error)
}

type Expander interface {
	Expand(expand *Expand, actingUserID, userID, channelID string) (*Expanded, error)
}

type service struct {
	Config     configurator.Service
	Expander   Expander
	Mattermost *pluginapi.Client

	Proxy         Proxy
	Registry      Registry
	Subscriptions Subscriptions
}

func NewService(mm *pluginapi.Client, configurator configurator.Service) Service {
	registry := NewRegistry(configurator)
	expander := NewExpander(mm, configurator)
	subs := NewSubscriptions(configurator)
	proxy := NewProxy(mm, configurator, subs)

	return &service{
		Config:        configurator,
		Expander:      expander,
		Mattermost:    mm,
		Proxy:         proxy,
		Registry:      registry,
		Subscriptions: subs,
	}
}
