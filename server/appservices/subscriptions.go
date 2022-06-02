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
	err := r.Check(
		r.RequireActingUser,
		r.RequireSourceApp,
		sub.Validate,
		a.canSubscribe(r, sub),
	)
	if err != nil {
		return err
	}

	// If there was a prior same-scoped subscription from the app, remove it.
	all, err := a.unsubscribe(r, sub.Event)
	if err != nil && errors.Cause(err) != utils.ErrNotFound {
		return err
	}

	// Make and save the new subscription.
	all = append(all, store.Subscription{
		Call:        sub.Call,
		AppID:       r.SourceAppID(),
		OwnerUserID: r.ActingUserID(),
	})
	err = a.store.Subscription.Save(sub.Event, all)
	if err != nil {
		return err
	}

	r.Log.Debugf("subscribed %s to %v, stored %v subscriptions", r.SourceAppID(), sub.Event, len(all))
	return nil
}

func (a *AppServices) Unsubscribe(r *incoming.Request, e apps.Event) error {
	err := r.Check(
		r.RequireActingUser,
		r.RequireSourceApp,
		e.Validate,
	)
	if err != nil {
		return err
	}

	modified, err := a.unsubscribe(r, e)
	if err != nil {
		return err
	}
	r.Log.Debugf("unsubscribed %s from %v, stored %v subscriptions", r.SourceAppID(), e, len(modified))

	return err
}

func (a *AppServices) GetSubscriptions(r *incoming.Request) (out []apps.Subscription, err error) {
	if err = r.Check(
		r.RequireActingUser,
		r.RequireSourceApp,
	); err != nil {
		return nil, err
	}

	allStored, err := a.store.Subscription.List()
	if err != nil {
		return nil, err
	}

	for _, stored := range allStored {
		for _, s := range stored.Subscriptions {
			if s.AppID == r.SourceAppID() && s.OwnerUserID == r.ActingUserID() {
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
	all, err := a.store.Subscription.Get(e)
	if err != nil {
		return nil, err
	}

	for i, s := range all {
		if s.AppID != r.SourceAppID() || s.OwnerUserID != r.ActingUserID() {
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

func (a *AppServices) canSubscribe(r *incoming.Request, sub apps.Subscription) func() error {
	return func() error {
		mm := r.Config().MattermostAPI()
		userID := r.ActingUserID()

		switch sub.Subject {
		case apps.SubjectUserCreated:
			if !mm.User.HasPermissionTo(userID, model.PermissionViewMembers) {
				return errors.New("no permission to read user")
			}

		case apps.SubjectUserJoinedChannel,
			apps.SubjectUserLeftChannel, /*, apps.SubjectPostCreated */
			apps.SubjectBotJoinedChannel,
			apps.SubjectBotLeftChannel /*, apps.SubjectBotMentioned*/ :
			if !mm.User.HasPermissionToChannel(userID, sub.ChannelID, model.PermissionReadChannel) {
				return errors.New("no permission to read channel")
			}

		case apps.SubjectUserJoinedTeam,
			apps.SubjectUserLeftTeam,
			apps.SubjectBotJoinedTeam,
			apps.SubjectBotLeftTeam:
			if !mm.User.HasPermissionToTeam(userID, sub.TeamID, model.PermissionViewTeam) {
				return errors.New("no permission to view team")
			}

		case apps.SubjectChannelCreated:
			if !mm.User.HasPermissionToTeam(userID, sub.TeamID, model.PermissionListTeamChannels) {
				return errors.New("no permission to list channels")
			}

		default:
			return errors.Errorf("Unknown subject %s", sub.Subject)
		}

		return nil
	}
}

// func SubscriptionsForAppUser(in []Subscription, appID apps.AppID, userID string) []Subscription {
// 	var out []Subscription
// 	for _, s := range in {
// 		if s.AppID == appID && s.CreatedByUserID == userID {
// 			out = append(out, s)
// 		}
// 	}
// 	return out
// }
