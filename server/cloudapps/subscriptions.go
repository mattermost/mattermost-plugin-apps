// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package cloudapps

import (
	"github.com/mattermost/mattermost-plugin-cloudapps/server/configurator"
)

type Subscriptions interface {
	GetSubscriptionsForChannel(subj SubscriptionSubject, channelID string) ([]*Subscription, error)
}

type subscriptions struct {
	configurator.Configurator
}

var _ Subscriptions = (*subscriptions)(nil)

func NewSubscriptions(configurator configurator.Configurator) Subscriptions {
	return &subscriptions{
		Configurator: configurator,
	}
}

func (subs *subscriptions) GetSubscriptionsForChannel(subj SubscriptionSubject, channelID string) ([]*Subscription, error) {
	return []*Subscription{
		{
			AppID:     "Hello",
			Subject:   SubjectUserJoinedChannel,
			ChannelID: "",
		},
	}, nil
}
