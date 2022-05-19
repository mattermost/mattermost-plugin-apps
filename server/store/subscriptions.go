// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type SubscriptionStore interface {
	Get(_ apps.Subject, teamID, channelID string) ([]apps.Subscription, error)
	List() ([]apps.Subscription, error)
	ListByUserID(_ apps.AppID, userID string) ([]apps.Subscription, error)
	Save(apps.Subscription) error
	Delete(apps.Subscription) error
}

type subscriptionStore struct {
	*Service
}

var _ SubscriptionStore = (*subscriptionStore)(nil)

func subsKey(subject apps.Subject, teamID, channelID string) (string, error) {
	idSuffix := ""
	switch subject {
	case apps.SubjectUserCreated,
		apps.SubjectBotJoinedChannel,
		apps.SubjectBotLeftChannel,
		apps.SubjectBotJoinedTeam,
		apps.SubjectBotLeftTeam /*, apps.SubjectBotMentioned */ :
		// Global, no suffix
	case apps.SubjectUserJoinedChannel,
		apps.SubjectUserLeftChannel /* , apps.SubjectPostCreated */ :
		idSuffix = "." + channelID
	case apps.SubjectUserJoinedTeam,
		apps.SubjectUserLeftTeam,
		apps.SubjectChannelCreated:
		idSuffix = "." + teamID
	default:
		return "", errors.Errorf("Unknown subject %s", subject)
	}
	return config.KVSubPrefix + string(subject) + idSuffix, nil
}

func (s subscriptionStore) Get(subject apps.Subject, teamID, channelID string) ([]apps.Subscription, error) {
	key, err := subsKey(subject, teamID, channelID)
	if err != nil {
		return nil, err
	}

	var subs []apps.Subscription
	err = s.conf.MattermostAPI().KV.Get(key, &subs)
	if err != nil {
		return nil, err
	}
	if len(subs) == 0 {
		return nil, utils.ErrNotFound
	}
	return subs, nil
}

func (s subscriptionStore) List() ([]apps.Subscription, error) {
	keys, err := s.conf.MattermostAPI().KV.ListKeys(0, keysPerPage, pluginapi.WithPrefix(config.KVSubPrefix))
	if err != nil {
		return nil, err
	}

	subs := []apps.Subscription{}
	for _, key := range keys {
		sub := []apps.Subscription{}
		err := s.conf.MattermostAPI().KV.Get(key, &sub)
		if err != nil {
			return nil, err
		}

		subs = append(subs, sub...)
	}
	return subs, nil
}

func (s subscriptionStore) ListByUserID(appID apps.AppID, userID string) ([]apps.Subscription, error) {
	subs, err := s.List()
	if err != nil {
		return nil, err
	}

	var rSubs []apps.Subscription
	for _, s := range subs {
		if s.AppID == appID && s.UserID == userID {
			rSubs = append(rSubs, s)
		}
	}

	return rSubs, nil
}

func (s subscriptionStore) Save(sub apps.Subscription) error {
	key, err := subsKey(sub.Subject, sub.TeamID, sub.ChannelID)
	if err != nil {
		return err
	}
	// get all subscriptions for the subject
	var subs []apps.Subscription
	err = s.conf.MattermostAPI().KV.Get(key, &subs)
	if err != nil {
		return err
	}

	add := true
	for i, s := range subs {
		if s.EqualScope(sub) {
			subs[i] = sub
			add = false
			break
		}
	}
	if add {
		subs = append(subs, sub)
	}

	_, err = s.conf.MattermostAPI().KV.Set(key, subs)
	if err != nil {
		return err
	}
	return nil
}

func (s subscriptionStore) Delete(sub apps.Subscription) error {
	key, err := subsKey(sub.Subject, sub.TeamID, sub.ChannelID)
	if err != nil {
		return err
	}

	// get all subscriptions for the subject
	var subs []apps.Subscription
	err = s.conf.MattermostAPI().KV.Get(key, &subs)
	if err != nil {
		return err
	}

	for i, current := range subs {
		if !sub.EqualScope(current) {
			continue
		}

		// sub exists and needs to be deleted
		updated := subs[:i]
		if i < len(subs) {
			updated = append(updated, subs[i+1:]...)
		}

		_, err = s.conf.MattermostAPI().KV.Set(key, updated)
		if err != nil {
			return errors.Wrap(err, "failed to save subscriptions")
		}
		return nil
	}

	return utils.ErrNotFound
}
