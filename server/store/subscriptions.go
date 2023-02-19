// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Subscription struct {
	Call        apps.Call  `json:"call"`
	AppID       apps.AppID `json:"app_id"`
	OwnerUserID string     `json:"user_id"`
}

type Subscriptions []Subscription

func (s Subscriptions) Clone() *Subscriptions {
	out := make(Subscriptions, len(s))
	copy(out, s)
	return &out
}

type SubscriptionStore struct {
	CachedStore[Subscriptions]
}

func (s *Service) makeSubscriptionStore(log utils.Logger) (*SubscriptionStore, error) {
	cached, err := MakeCachedStore[Subscriptions](SubscriptionStoreName, s.cluster, log)
	if err != nil {
		return nil, err
	}
	return &SubscriptionStore{
		CachedStore: cached,
	}, nil
}

func (s *SubscriptionStore) Get(_ *incoming.Request, e apps.Event) (Subscriptions, error) {
	key := utils.ToJSON(e)
	if key == "{}" {
		return nil, errors.New("failed to get subscriptions: invalid empty event")
	}
	subs := s.CachedStore.Get(key)
	if subs == nil {
		return nil, errors.Wrapf(utils.ErrNotFound, "failed to get subscriptions for event %s", e.String())
	}
	return *subs, nil
}

func (s *SubscriptionStore) ListSubscribedEvents(_ *incoming.Request) ([]apps.Event, error) {
	var events []apps.Event
	for eventJSON := range s.Index() {
		var e apps.Event
		if err := json.Unmarshal([]byte(eventJSON), &e); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (s *SubscriptionStore) Put(r *incoming.Request, e apps.Event, subs *Subscriptions) error {
	key := utils.ToJSON(e)
	if key == "{}" {
		return errors.New("failed to save subscriptions: invalid empty event")
	}
	return s.CachedStore.Put(r, key, subs)
}

func (s Subscription) String() string {
	return fmt.Sprintf("%s/%s/%s", s.AppID, s.OwnerUserID, s.Call.Path)
}
