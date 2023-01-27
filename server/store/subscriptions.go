// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"encoding/json"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/plugin"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Subscription struct {
	Call        apps.Call  `json:"call"`
	AppID       apps.AppID `json:"app_id"`
	OwnerUserID string     `json:"user_id"`
}

type SubscriptionStore struct {
	*CachedStore[[]Subscription]
}

func MakeSubscriptionStore(api plugin.API, mmapi *pluginapi.Client, log utils.Logger) (*SubscriptionStore, error) {
	store, err := MakeCachedStore[[]Subscription](SubscriptionStoreName, api, mmapi, log)
	if err != nil {
		return nil, err
	}
	return &SubscriptionStore{
		CachedStore: store,
	}, nil
}

func (s *SubscriptionStore) Get(_ *incoming.Request, e apps.Event) ([]Subscription, error) {
	key := utils.ToJSON(e)
	if key == "{}" {
		return nil, errors.New("failed to get subscriptions: invalid empty event")
	}
	subs, ok := s.GetCachedStoreItem(key)
	if !ok {
		return nil, errors.Wrapf(utils.ErrNotFound, "failed to get subscriptions for event %s", key)
	}
	return subs, nil
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

func (s *SubscriptionStore) Save(r *incoming.Request, e apps.Event, subs []Subscription) error {
	key := utils.ToJSON(e)
	if key == "{}" {
		return errors.New("failed to get subscriptions: invalid empty event")
	}
	return s.PutCachedStoreItem(r, key, subs)
}
