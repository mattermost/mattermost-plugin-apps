// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package appservices

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type PermissionChecker interface {
	HasPermissionTo(userID string, permission *model.Permission) bool
	HasPermissionToChannel(userID, channelID string, permission *model.Permission) bool
	HasPermissionToTeam(userID, teamID string, permission *model.Permission) bool
}

func (a *AppServices) Subscribe(r *incoming.Request, sub apps.Subscription) error {
	if err := sub.Validate(); err != nil {
		return err
	}
	if err := a.canSubscribe(r, sub); err != nil {
		return err
	}

	// If there was a prior same-scoped subscription from the app, remove it.
	all, err := a.unsubscribe(r, sub.Event)
	if err != nil && errors.Cause(err) != utils.ErrNotFound {
		return err
	}
	all = append(all, store.Subscription{
		Call:            sub.Call,
		AppID:           r.SourceAppID(),
		CreatedByUserID: r.ActingUserID(),
	})
	return a.store.Subscription.Save(sub.Event, all)
}

func (a *AppServices) Unsubscribe(r *incoming.Request, e apps.Event) error {
	_, err := a.unsubscribe(r, e)
	return err
}

func (a *AppServices) GetSubscriptions(r *incoming.Request) (out []apps.Subscription, err error) {
	allStored, err := a.store.Subscription.List()
	if err != nil {
		return nil, err
	}

	for _, stored := range allStored {
		for _, s := range stored.Subscriptions {
			if s.AppID == r.SourceAppID() {
				out = append(out, apps.Subscription{
					Event: stored.Event,
					Call:  s.Call,
				})
			}
		}
	}

	return out, nil
}

func (a *AppServices) unsubscribe(r *incoming.Request, e apps.Event) ([]store.Subscription, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	all, err := a.store.Subscription.Get(e)
	if err != nil {
		return nil, err
	}

	appID := r.SourceAppID()
	for i, s := range all {
		if s.AppID != appID {
			continue
		}

		// TODO: check permissions to unsubscribe. For now, all subscribe APIs
		// are sysadmin-only, so nothing to check.

		modified := all[:i]
		if i < len(all) {
			modified = append(modified, all[i+1:]...)
		}
		err = a.store.Subscription.Save(e, modified)
		if err != nil {
			return nil, err
		}
		return modified, nil
	}

	return all, utils.ErrNotFound
}

func (a *AppServices) canSubscribe(r *incoming.Request, sub apps.Subscription) error {
	userID := sub.CreatedByUserID

	switch sub.Subject {
	case apps.SubjectUserCreated:
		if !checker.HasPermissionTo(userID, model.PermissionViewMembers) {
			return errors.New("no permission to read user")
		}
	case apps.SubjectUserJoinedChannel, apps.SubjectUserLeftChannel /*, apps.SubjectPostCreated */ :
		if !checker.HasPermissionToChannel(userID, sub.ChannelID, model.PermissionReadChannel) {
			return errors.New("no permission to read channel")
		}
	case apps.SubjectBotJoinedChannel, apps.SubjectBotLeftChannel /*, apps.SubjectBotMentioned*/ :
		// Only check if there is dynamic context i.e. a channelID
		if channelID != "" {
			if !checker.HasPermissionToChannel(userID, channelID, model.PermissionReadChannel) {
				return errors.New("no permission to read channel")
			}
		}
	case apps.SubjectUserJoinedTeam,
		apps.SubjectUserLeftTeam:
		if !checker.HasPermissionToTeam(userID, sub.TeamID, model.PermissionViewTeam) {
			return errors.New("no permission to view team")
		}
	case apps.SubjectBotJoinedTeam,
		apps.SubjectBotLeftTeam:
		// Only check if there is dynamic context i.e. a channelID
		if teamID != "" {
			if !checker.HasPermissionToTeam(userID, teamID, model.PermissionViewTeam) {
				return errors.New("no permission to read channel")
			}
		}
	case apps.SubjectChannelCreated:
		if !checker.HasPermissionToTeam(userID, sub.TeamID, model.PermissionListTeamChannels) {
			return errors.New("no permission to list channels")
		}
	default:
		return errors.Errorf("Unknown subject %s", sub.Subject)
	}

	return nil
}

func SubscriptionsForAppUser(in []Subscription, appID apps.AppID, userID string) []Subscription {
	var out []Subscription
	for _, s := range in {
		if s.AppID == appID && s.CreatedByUserID == userID {
			out = append(out, s)
		}
	}
	return out
}

func (s subscriptionStore) Save(sub Subscription) error {
	key, err := subsKey(sub.Subject, sub.TeamID, sub.ChannelID)
	if err != nil {
		return err
	}
	// get all subscriptions for the subject
	var subs []Subscription
	err = s.conf.MattermostAPI().KV.Get(key, &subs)
	if err != nil {
		return err
	}

	add := true
	for i, s := range subs {
		if s.EqualScope(sub) && s.AppID == sub.AppID {
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

func (s subscriptionStore) Delete(sub Subscription) error {
	key, err := subsKey(sub.Subject, sub.TeamID, sub.ChannelID)
	if err != nil {
		return err
	}

	// get all subscriptions for the subject
	var subs []Subscription
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

func (s Subscription) EqualScope(other Subscription) bool {
	return s.AppID == other.AppID && s.Subscription.EqualScope(other.Subscription)
}
