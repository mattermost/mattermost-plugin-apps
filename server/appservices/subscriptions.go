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
		a.hasPermissionToSubscribe(r, sub),
	)
	if err != nil {
		return err
	}

	// If there was a prior same-scoped subscription from the app, remove it.
	ownerID := r.ActingUserID()
	all, err := a.unsubscribe(r, ownerID, sub.Event)
	if err != nil && errors.Cause(err) != utils.ErrNotFound {
		return err
	}

	// Make and save the new subscription.
	all = append(all, store.Subscription{
		Call:        sub.Call,
		AppID:       r.SourceAppID(),
		OwnerUserID: ownerID,
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

	modified, err := a.unsubscribe(r, r.ActingUserID(), e)
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

func (a *AppServices) UnsubscribeApp(r *incoming.Request, appID apps.AppID) error {
	err := r.Check(
		r.RequireSysadminOrPlugin,
	)
	if err != nil {
		return err
	}

	allStored, err := a.store.Subscription.List()
	if err != nil {
		return err
	}

	n := 0
	for _, stored := range allStored {
		modified := []store.Subscription{}
		for _, s := range stored.Subscriptions {
			if s.AppID == appID {
				n++
			} else {
				modified = append(modified, s)
			}
		}
		if len(modified) < len(stored.Subscriptions) {
			err = a.store.Subscription.Save(stored.Event, modified)
			if err != nil {
				return err
			}
		}
	}

	r.Log.Debugf("removed all (%v) subscriptions for %s", n, appID)
	return err
}

func (a *AppServices) unsubscribe(r *incoming.Request, ownerUserID string, e apps.Event) ([]store.Subscription, error) {
	all, err := a.store.Subscription.Get(e)
	if err != nil {
		return nil, err
	}

	for i, s := range all {
		if s.AppID != r.SourceAppID() || s.OwnerUserID != ownerUserID {
			continue
		}

		allowed := false
		if s.OwnerUserID == r.ActingUserID() {
			allowed = true
		} else if err = r.RequireSysadminOrPlugin(); err == nil {
			allowed = true
		}
		if !allowed {
			return nil, utils.NewForbiddenError("must be the owner of the subscription, or system administrator to unsubscribe")
		}

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

	return all, errors.Wrap(utils.ErrNotFound, "You are not subscribed to this notification")
}

func (a *AppServices) hasPermissionToSubscribe(r *incoming.Request, sub apps.Subscription) func() error {
	return func() error {
		mm := r.Config().MattermostAPI()
		userID := r.ActingUserID()

		switch sub.Subject {
		case apps.SubjectUserCreated:
			if !mm.User.HasPermissionTo(userID, model.PermissionViewMembers) {
				return errors.New("no permission to read user")
			}

		case apps.SubjectUserJoinedChannel, apps.SubjectUserLeftChannel /*, apps.SubjectPostCreated, apps.SubjectBotMentioned */ :
			if !mm.User.HasPermissionToChannel(userID, sub.ChannelID, model.PermissionReadChannel) {
				return errors.New("no permission to read channel")
			}

		case apps.SubjectUserJoinedTeam, apps.SubjectUserLeftTeam:
			if !mm.User.HasPermissionToTeam(userID, sub.TeamID, model.PermissionViewTeam) {
				return errors.New("no permission to view team")
			}

		case apps.SubjectBotJoinedChannel,
			apps.SubjectBotLeftChannel,
			apps.SubjectBotJoinedTeam,
			apps.SubjectBotLeftTeam:
			// When the bot has joined an entity, it will have the permission to
			// read it.
			return nil

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
