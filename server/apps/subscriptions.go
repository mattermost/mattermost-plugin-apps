// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
)

const SubsPrefixKey = "sub"

type Subscriptions interface {
	GetSubscriptionsForChannel(subj SubscriptionSubject, channelID string) ([]*Subscription, error)
	GetSubscriptionsForTeam(subj SubscriptionSubject, teamID string) ([]*Subscription, error)
	GetSubscriptionsForApp(appID string, subj SubscriptionSubject, teamID string) ([]*Subscription, error)
	StoreSubcription(subj SubscriptionSubject, channelID string) error
	DeleteSubcription(subj SubscriptionSubject, channelID string) error
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

// GetSubscriptionsForChannel returns subscriptions for a given subject and
// channelID from the store
func (subs *subscriptions) GetSubscriptionsForChannel(subj SubscriptionSubject, channelID string) ([]*Subscription, error) {
	// TODO Implement KVGet
	//   check if key subs_channelID exists
	//   if yes, return them
	//   if no, return empty (or nil??) subs
	return []*Subscription{
		{
			AppID:     "Hello",
			Subject:   SubjectUserJoinedChannel,
			ChannelID: "",
		},
	}, nil
}

func (subs *subscriptions) GetSubscriptionsForTeam(subj SubscriptionSubject, channelID string) ([]*Subscription, error) {
	return nil, nil
}

func (subs *subscriptions) GetSubscriptionsForApp(app string, subj SubscriptionSubject, channelID string) ([]*Subscription, error) {
	// if subj is nil, grab all subjects for the
	return nil, nil
}

// StoreSubscription stores a subscription
func (subs *subscriptions) StoreSubcription(subj SubscriptionSubject, channelID string) error {
	// TODO Implement KVSet
	//   check if subscription already exists
	//   if yes, update/overwrite
	//   if no, append new subscription to subs_channelID values
	return nil
}

// DeleteSubscription deletes a subscription
func (subs *subscriptions) DeleteSubcription(subj SubscriptionSubject, channelID string) error {
	// TODO Implement KVDelete
	//	  check if subscription exists
	//    if yes, delete from subs_channelID values
	//    if no, return nil
	return nil
}
