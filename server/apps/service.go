// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
)

type AppClient interface {
	PostWish(toAppID AppID, fromMattermostUserID string, w *Wish, data *CallData) (*CallResponse, error)
	PostChangeNotification(AppID, *Subscription, interface{})
}

type Hooks interface {
	OnUserJoinedChannel(pluginContext *plugin.Context, channelMember *model.ChannelMember, actor *model.User)
}

type API interface {
	Call(AppID, string, *Call) (*CallResponse, error)
	InstallApp(*InInstallApp) (*OutInstallApp, error)
}

type Subscriptions interface {
	GetSubscriptionsForChannel(subj SubscriptionSubject, channelID string) ([]*Subscription, error)
}

type Registry interface {
	Store(*App) error
	Get(AppID) (*App, error)
}

type Expander interface {
	Expand(expand *Expand, actingUserID, userID, channelID string) (*Expanded, error)
}

type Service struct {
	Configurator  configurator.Service
	Mattermost    *pluginapi.Client
	Expander      Expander
	Registry      Registry
	Subscriptions Subscriptions
	AppClient     AppClient
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
	s.AppClient = s
	s.API = s

	return s
}
