// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
	"github.com/pkg/errors"
)

const SubsPrefixKey = "sub"

type Subscriptions interface {
	GetSubscriptionsForChannel(subj SubscriptionSubject, channelID string) ([]*Subscription, error)
	GetSubscriptionsForTeam(subj SubscriptionSubject, teamID string) ([]*Subscription, error)
	GetSubscriptionsForApp(appID string, subj SubscriptionSubject, teamID string) ([]*Subscription, error)
	StoreSubscription(subj SubscriptionSubject, sub Subscription, channelID string) error
	DeleteSubscription(subj SubscriptionSubject, sub SubscriptionID, channelID string) error
}

type subscriptions struct {
	configurator.Configurator
	mm *pluginapi.Client
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
			Subject:   subj,
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

// StoreSubscription stores a subscription for a change notification
// TODO move this to store package or file
func (subs *subscriptions) StoreSubscription(subj SubscriptionSubject, sub Subscription, channelID string) error {
	key := GetSubsKVkey(subj, channelID)

	// get all subscriptions for the subject
	var savedSubs []*Subscription
	if err := subs.mm.KV.Get(key, &savedSubs); err != nil {
		return errors.Wrap(err, "failed to get saved subscriptions")
	}

	// check if sub exists
	var newSubs []*Subscription
	for _, s := range savedSubs {
		// modify the sub to the latest request
		if s.SubscriptionID == sub.SubscriptionID {
			newSubs = append(newSubs, &sub)
			continue
		}
		newSubs = append(newSubs, s)
	}

	// sub exists. update and save updated subs
	_, err := subs.mm.KV.Set(key, newSubs)
	if err != nil {
		return errors.Wrap(err, "failed to save subscriptions")
	}
	return nil
}

// DeleteSubscription deletes a subscription
func (subs *subscriptions) DeleteSubscription(subj SubscriptionSubject, subID SubscriptionID, channelID string) error {
	key := GetSubsKVkey(subj, channelID)

	// get all subscriptions for the subject
	var savedSubs []*Subscription
	if err := subs.mm.KV.Get(key, &savedSubs); err != nil {
		return errors.Wrap(err, "failed to get saved subscriptions")
	}

	// check if sub exists
	var newSubs []*Subscription
	for _, s := range savedSubs {
		if s.SubscriptionID == subID {
			continue
		}
		newSubs = append(newSubs, s)
	}

	// sub was deleted. update and save updated subs
	// TODO check for don't need to save if sub was not deleted?
	_, err := subs.mm.KV.Set(key, newSubs)
	if err != nil {
		return errors.Wrap(err, "failed to save subscriptions")
	}
	return nil
}

func GetSubsKVkey(subj SubscriptionSubject, channelID string) string {
	key := SubsPrefixKey
	switch subj {
	case SubjectUserJoinedChannel:
		key += "_" + string(SubjectUserJoinedChannel) + channelID
	// case value2:
	default:
	}
	return key
}
