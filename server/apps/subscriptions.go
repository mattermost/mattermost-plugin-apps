// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"fmt"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/pkg/errors"
)

const SubsPrefixKey = "sub_"

type Subscriptions interface {
	GetChannelOrTeamSubs(subj constants.SubscriptionSubject, channelOrTeamID string) ([]*Subscription, error)
	GetAppSubs(appID string, subj constants.SubscriptionSubject, teamID string) ([]*Subscription, error)
	StoreSub(sub *Subscription) error
	DeleteSub(sub *Subscription) error
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
func (s *subscriptions) GetChannelOrTeamSubs(subj constants.SubscriptionSubject, channelOrTeamID string) ([]*Subscription, error) {
	key, err := s.getAndValidateSubsKVkey(nil, subj, channelOrTeamID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get subscriptions key")
	}
	var savedSubs []*Subscription
	if err := s.mm.KV.Get(key, &savedSubs); err != nil {
		return nil, errors.Wrap(err, "failed to get saved subscriptions")
	}
	if len(savedSubs) == 0 {
		return nil, utils.ErrNotFound
	}

	return savedSubs, nil
}

func (s *subscriptions) GetAppSubs(app string, subj constants.SubscriptionSubject, channelID string) ([]*Subscription, error) {
	// if subj is nil, grab all subjects for the
	return nil, nil
}

// StoreSub stores a subscription for a change notification
// TODO move this to store package or file
func (s *subscriptions) StoreSub(sub *Subscription) error {
	if sub.Subject == "" {
		return errors.New("failed to get subscription subject")
	}
	key, err := s.getAndValidateSubsKVkey(sub, "", "")
	if err != nil {
		return errors.Wrap(err, "failed to get subscriptions key")
	}

	// get all subscriptions for the subject
	var subs []*Subscription
	if err = s.mm.KV.Get(key, &subs); err != nil {
		return errors.Wrap(err, "failed to get subscriptions")
	}

	add := true
	for i, s := range subs {
		if s.SubscriptionID == sub.SubscriptionID {
			subs[i] = sub
			add = false
			break
		}
	}
	if add {
		subs = append(subs, sub)
	}

	_, err = s.mm.KV.Set(key, subs)
	if err != nil {
		return errors.Wrap(err, "failed to save subscriptions")
	}
	return nil
}

// DeleteSubs deletes a subscription
func (s *subscriptions) DeleteSub(sub *Subscription) error {
	if sub.Subject == "" {
		return errors.New("failed to get subscription subject")
	}
	if sub.ChannelID == "" {
		return errors.New("failed to get subscription channelID")
	}

	key, err := s.getAndValidateSubsKVkey(sub, sub.Subject, sub.ChannelID)
	if err != nil {
		return errors.Wrap(err, "failed to get subscriptions key")
	}

	// get all subscriptions for the subject
	var subs []*Subscription
	if err = s.mm.KV.Get(key, &subs); err != nil {
		return errors.Wrap(err, "failed to get saved subscriptions")
	}

	for i, current := range subs {
		if current.SubscriptionID != sub.SubscriptionID {
			continue
		}

		// sub exists and dees to be deleted
		updated := subs[:i]
		if i < len(subs) {
			updated = append(updated, subs[i+1:]...)
		}

		_, err = s.mm.KV.Set(key, updated)
		if err != nil {
			return errors.Wrap(err, "failed to save subscriptions")
		}

		return nil
	}

	return utils.ErrNotFound
}

// GetSubsKey returns the KVstore Key for a subject. If teamOrChannelID
// provided, the value is appended to the subject key making it unique to the
// channelID or teamID. Also validates the team or channel ID exists
// TODO what to do if the app wants to delete a subscription for a channel that
// was deleted?
func (s *subscriptions) getAndValidateSubsKVkey(sub *Subscription, subject constants.SubscriptionSubject, teamOrChannelID string) (string, error) {
	if sub != nil {
		subject = sub.Subject
	}

	// verify valid subject request and create the key
	key := SubsPrefixKey + string(subject)
	switch subject {
	case constants.SubjectUserJoinedChannel,
		constants.SubjectUserLeftChannel:
		if sub != nil {
			teamOrChannelID = sub.ChannelID
		}
		_, errChan := s.mm.Channel.Get(teamOrChannelID)
		if errChan != nil {
			return "", errors.New(fmt.Sprintf("ChannelID %s does not exist", teamOrChannelID))
		}
		key += "_" + teamOrChannelID
	case constants.SubjectUserJoinedTeam,
		constants.SubjectUserLeftTeam:
		if sub != nil {
			teamOrChannelID = sub.TeamID
		}
		_, errTeam := s.mm.Team.Get(teamOrChannelID)
		if errTeam != nil {
			return "", errors.New(fmt.Sprintf("TeamID %s does not exist", teamOrChannelID))
		}
		key += "_" + teamOrChannelID
	case constants.SubjectChannelCreated,
		constants.SubjectPostCreated,
		constants.SubjectUserCreated,
		constants.SubjectUserUpdated:
	default:
		return "", errors.New(fmt.Sprintf("subject %s is not a valid subject", subject))
	}
	return key, nil
}
