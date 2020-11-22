// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/pkg/errors"
)

func (s *Store) subsKey(subject api.Subject, teamID, channelID string) string {
	idSuffix := ""
	switch subject {
	case api.SubjectUserJoinedChannel,
		api.SubjectUserLeftChannel,
		api.SubjectPostCreated:
		idSuffix = "_" + channelID
	case api.SubjectUserJoinedTeam,
		api.SubjectUserLeftTeam,
		api.SubjectChannelCreated:
		idSuffix = "_" + teamID
	}
	return prefixSubs + string(subject) + idSuffix
}

func (s *Store) DeleteSub(sub *api.Subscription) error {
	key := s.subsKey(sub.Subject, sub.TeamID, sub.ChannelID)
	// get all subscriptions for the subject
	var subs []*api.Subscription
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

func (s *Store) LoadSubs(subject api.Subject, teamID, channelID string) ([]*api.Subscription, error) {
	key := s.subsKey(subject, teamID, channelID)
	var subs []*api.Subscription
	err := s.mm.KV.Get(key, &subs)
	if err != nil {
		return nil, err
	}
	if len(subs) == 0 {
		return nil, utils.ErrNotFound
	}
	return subs, nil
}

func (s *Store) StoreSub(sub *api.Subscription) error {
	key := s.subsKey(sub.Subject, sub.TeamID, sub.ChannelID)
	// get all subscriptions for the subject
	var subs []*api.Subscription
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
