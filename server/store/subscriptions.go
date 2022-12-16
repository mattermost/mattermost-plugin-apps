// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// SubscriptionStore stores the complete (for all apps) list of subscriptions
// for each "scope", everything in apps.Subscription, but the Call - the
// subject, and the optional team/channel IDs.
type SubscriptionStore interface {
	Get(apps.Event) ([]Subscription, error)
	List() ([]StoredSubscriptions, error)
	Save(apps.Event, []Subscription) error
}

type Subscription struct {
	Call        apps.Call
	AppID       apps.AppID `json:"app_id"`
	OwnerUserID string     `json:"user_id"`
}

type StoredSubscriptions struct {
	Event         apps.Event
	Subscriptions []Subscription
}

type subscriptionStore struct {
	*Service
}

var _ SubscriptionStore = (*subscriptionStore)(nil)

func subsKey(e apps.Event) (string, error) {
	idSuffix := ""
	switch e.Subject {
	case apps.SubjectUserCreated,
		// apps.SubjectSelfMentioned,
		apps.SubjectBotJoinedTeam,
		apps.SubjectBotLeftTeam,
		apps.SubjectBotJoinedChannel,
		apps.SubjectBotLeftChannel,
		apps.SubjectSelfJoinedTeam,
		apps.SubjectSelfLeftTeam,
		apps.SubjectSelfJoinedChannel,
		apps.SubjectSelfLeftChannel:
	// Global subscriptions, no suffix

	case apps.SubjectUserJoinedChannel,
		apps.SubjectUserLeftChannel /* , apps.SubjectPostCreated */ :
		idSuffix = "." + e.ChannelID

	case apps.SubjectUserJoinedTeam,
		apps.SubjectUserLeftTeam,
		apps.SubjectChannelCreated:
		idSuffix = "." + e.TeamID
	default:
		return "", errors.Errorf("Unknown subject %s", e.Subject)
	}

	return KVSubPrefix + string(e.Subject) + idSuffix, nil
}

func (s subscriptionStore) Get(e apps.Event) ([]Subscription, error) {
	key, err := subsKey(e)
	if err != nil {
		return nil, err
	}

	stored := &StoredSubscriptions{}
	err = s.conf.MattermostAPI().KV.Get(key, &stored)
	if err != nil {
		return nil, err
	}
	if stored == nil {
		return nil, utils.ErrNotFound
	}
	return stored.Subscriptions, nil
}

func (s subscriptionStore) List() ([]StoredSubscriptions, error) {
	all := []StoredSubscriptions{}
	for i := 0; ; i++ {
		keys, err := s.conf.MattermostAPI().KV.ListKeys(i, ListKeysPerPage)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list keys - page, %d", i)
		}
		if len(keys) == 0 {
			break
		}

		for _, key := range keys {
			if !strings.HasPrefix(key, KVSubPrefix) {
				continue
			}
			forKey := StoredSubscriptions{}
			err := s.conf.MattermostAPI().KV.Get(key, &forKey)
			if err != nil {
				return nil, err
			}
			if forKey.Event.Subject == "" {
				continue
			}
			all = append(all, forKey)
		}
	}
	return all, nil
}

func (s subscriptionStore) Save(e apps.Event, subs []Subscription) error {
	key, err := subsKey(e)
	if err != nil {
		return err
	}

	if len(subs) == 0 {
		return s.conf.MattermostAPI().KV.Delete(key)
	}

	_, err = s.conf.MattermostAPI().KV.Set(key, StoredSubscriptions{
		Event:         e,
		Subscriptions: subs,
	})
	return err
}
