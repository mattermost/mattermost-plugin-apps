// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/server/appmodel"
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
)

type Subscriptions interface {
	GetSubscriptionsForChannel(subj appmodel.SubscriptionSubject, channelID string) ([]*appmodel.Subscription, error)
}

type subscriptions struct {
	configurator configurator.Service
}

var _ Subscriptions = (*subscriptions)(nil)

func NewSubscriptions(configurator configurator.Service) Subscriptions {
	return &subscriptions{
		configurator: configurator,
	}
}

func (subs *subscriptions) GetSubscriptionsForChannel(subj appmodel.SubscriptionSubject, channelID string) ([]*appmodel.Subscription, error) {
	return []*appmodel.Subscription{
		{
			AppID:     "Hello",
			Subject:   appmodel.SubjectUserJoinedChannel,
			ChannelID: "",
		},
	}, nil
}
