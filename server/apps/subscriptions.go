// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/pkg/errors"
)

const SubsPrefixKey = "sub_"

type Subscriptions interface {
	GetChannelOrTeamSubs(subj SubscriptionSubject, channelOrTeamID string) ([]*Subscription, error)
	GetAppSubs(appID string, subj SubscriptionSubject, teamID string) ([]*Subscription, error)
	StoreSub(sub Subscription) error
	DeleteSub(sub Subscription) error
}

type subscriptions struct {
	configurator configurator.Service
	mm           *pluginapi.Client
}

var _ Subscriptions = (*subscriptions)(nil)

func NewSubscriptions(mm *pluginapi.Client, configurator configurator.Service) Subscriptions {
	return &subscriptions{
		mm:           mm,
		configurator: configurator,
	}
}

// GetSubsForChannelOrTeam returns subscriptions for a given subject and
// channelID or teamID from the store
func (s *subscriptions) GetChannelOrTeamSubs(subj SubscriptionSubject, channelOrTeamID string) ([]*Subscription, error) {
	key := GetSubsKVkey(subj, channelOrTeamID)
	var savedSubs []*Subscription
	if err := s.mm.KV.Get(key, &savedSubs); err != nil {
		return nil, errors.Wrap(err, "failed to get saved subscriptions")
	}
	if len(savedSubs) == 0 {
		return nil, utils.ErrNotFound
	}

	return savedSubs, nil
}

func (s *subscriptions) GetAppSubs(app string, subj SubscriptionSubject, channelID string) ([]*Subscription, error) {
	// if subj is nil, grab all subjects for the
	return nil, nil
}

// StoreSub stores a subscription for a change notification
// TODO move this to store package or file
func (s *subscriptions) StoreSub(sub Subscription) error {

	if sub.Subject == "" {
		return errors.New("failed to get subscription subject")
	}
	// TODO check that channelID exists
	if sub.ChannelID == "" {
		return errors.New("failed to get subscription channelID")
	}
	key := GetSubsKVkey(sub.Subject, sub.ChannelID)

	// get all subscriptions for the subject
	var savedSubs []*Subscription
	if err := s.mm.KV.Get(key, &savedSubs); err != nil {
		return errors.Wrap(err, "failed to get subscriptions")
	}

	// check if sub exists
	var newSubs []*Subscription
	foundSub := 0
	for i, s := range savedSubs {
		// modify the sub to the latest request
		if s.SubscriptionID == sub.SubscriptionID {
			foundSub++
			newSubs = append(newSubs, &sub)
			newSubs = append(newSubs, savedSubs[i+1:]...)
			break
		}
		newSubs = append(newSubs, s)
	}
	if foundSub == 0 {
		newSubs = append(newSubs, &sub)
	}

	// sub exists. update and save updated subs
	_, err := s.mm.KV.Set(key, newSubs)
	if err != nil {
		return errors.Wrap(err, "failed to save subscriptions")
	}
	return nil
}

// DeleteSubs deletes a subscription
func (s *subscriptions) DeleteSub(sub Subscription) error {
	if sub.Subject == "" {
		return errors.New("failed to get subscription subject")
	}
	// TODO check that channelID exists
	if sub.ChannelID == "" {
		return errors.New("failed to get subscription channelID")
	}
	key := GetSubsKVkey(sub.Subject, sub.ChannelID)

	// get all subscriptions for the subject
	var savedSubs []*Subscription
	if err := s.mm.KV.Get(key, &savedSubs); err != nil {
		return errors.Wrap(err, "failed to saved subscriptions")
	}

	// check if sub exists
	var newSubs []*Subscription
	for i, s := range savedSubs {
		if s.SubscriptionID == sub.SubscriptionID {
			newSubs = append(newSubs, savedSubs[i+1:]...)
			break
		}
		newSubs = append(newSubs, s)
	}

	// sub was deleted. update and save updated subs
	// TODO check for following:
	//   - don't need to save if sub was not deleted?
	//   - if delete the last subscription for the channel, delete the key also
	_, err := s.mm.KV.Set(key, newSubs)
	if err != nil {
		return errors.Wrap(err, "failed to save subscriptions")
	}

	return nil
}

// GetSubsKey returns the KVstore Key for a subsject. If teamOrChannelID
// provided, the value is appended to the subject key making it unique to the
// channelID or teamID
func GetSubsKVkey(subj SubscriptionSubject, teamOrChannelID string) string {
	key := SubsPrefixKey + string(subj)
	switch subj {
	case SubjectUserJoinedChannel,
		SubjectUserLeftChannel,
		SubjectUserJoinedTeam,
		SubjectUserLeftTeam:
		key += "_" + teamOrChannelID
	// case value2:
	default:
	}
	return key
}
