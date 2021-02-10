// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

type SubStore struct {
	*Store
}

var _ api.SubStore = (*SubStore)(nil)

func newSubStore(st *Store) api.SubStore {
	s := &SubStore{st}
	return s
}

func subsKey(subject apps.Subject, teamID, channelID string) string {
	idSuffix := ""
	switch subject {
	case apps.SubjectUserJoinedChannel,
		apps.SubjectUserLeftChannel,
		apps.SubjectPostCreated:
		idSuffix = "_" + channelID
	case apps.SubjectUserJoinedTeam,
		apps.SubjectUserLeftTeam,
		apps.SubjectChannelCreated:
		idSuffix = "_" + teamID
	}
	return prefixSubs + string(subject) + idSuffix
}

func (s SubStore) Delete(sub *apps.Subscription) error {
	key := subsKey(sub.Subject, sub.TeamID, sub.ChannelID)
	// get all subscriptions for the subject
	var subs []*apps.Subscription
	err := s.mm.KV.Get(key, &subs)
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

		_, err = s.mm.KV.Set(key, updated)
		if err != nil {
			return errors.Wrap(err, "failed to save subscriptions")
		}
		return nil
	}

	return utils.ErrNotFound
}

func (s SubStore) Get(subject apps.Subject, teamID, channelID string) ([]*apps.Subscription, error) {
	key := subsKey(subject, teamID, channelID)
	var subs []*apps.Subscription
	err := s.mm.KV.Get(key, &subs)
	if err != nil {
		return nil, err
	}
	if len(subs) == 0 {
		return nil, utils.ErrNotFound
	}
	return subs, nil
}

func (s SubStore) Save(sub *apps.Subscription) error {
	key := subsKey(sub.Subject, sub.TeamID, sub.ChannelID)
	// get all subscriptions for the subject
	var subs []*apps.Subscription
	err := s.mm.KV.Get(key, &subs)
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

	_, err = s.mm.KV.Set(key, subs)
	if err != nil {
		return err
	}
	return nil
}
