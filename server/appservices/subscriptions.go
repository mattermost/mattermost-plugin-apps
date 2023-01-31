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
	err = a.subscriptions.Save(r, sub.Event, all)
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

	events, err := a.subscriptions.ListSubscribedEvents(r)
	if err != nil {
		return nil, err
	}
	for _, event := range events {
		var subs []store.Subscription
		subs, err = a.subscriptions.Get(r, event)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get subscriptions for "+event.String())
		}
		for _, s := range subs {
			if s.AppID == r.SourceAppID() && s.OwnerUserID == r.ActingUserID() {
				out = append(out, apps.Subscription{
					Event: event,
					Call:  s.Call,
				})
			}
		}
	}

	return out, nil
}

func (a *AppServices) unsubscribeApp(r *incoming.Request, appID apps.AppID) error {
	err := r.Check(
		r.RequireSysadminOrPlugin,
	)
	if err != nil {
		return err
	}

	events, err := a.subscriptions.ListSubscribedEvents(r)
	if err != nil {
		return err
	}

	n := 0
	for _, event := range events {
		var subs []store.Subscription
		modified := []store.Subscription{}
		subs, err = a.subscriptions.Get(r, event)
		if err != nil {
			return errors.Wrap(err, "failed to get subscriptions for "+event.String())
		}
		for _, sub := range subs {
			if sub.AppID == appID {
				n++
			} else {
				modified = append(modified, sub)
			}
		}
		if len(modified) < len(subs) {
			if len(modified) > 0 {
				err = a.subscriptions.Save(r, event, modified)
			} else {
				err = a.subscriptions.Delete(r, event)
			}
			if err != nil {
				return err
			}
		}
	}

	r.Log.Debugf("removed all (%v) subscriptions for %s", n, appID)
	return err
}

func (a *AppServices) unsubscribe(r *incoming.Request, ownerUserID string, event apps.Event) ([]store.Subscription, error) {
	all, err := a.subscriptions.Get(r, event)
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
		if len(modified) > 0 {
			err = a.subscriptions.Save(r, event, modified)
		} else {
			err = a.subscriptions.Delete(r, event)
		}
		if err != nil {
			return nil, err
		}

		return modified, nil
	}

	return all, errors.Wrapf(utils.ErrNotFound, "failed to get subscriptions for event %s", event.String())
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

		case apps.SubjectUserJoinedChannel, apps.SubjectUserLeftChannel:
			if sub.ChannelID != "" && !mm.User.HasPermissionToChannel(userID, sub.ChannelID, model.PermissionReadChannel) {
				return errors.New("no permission to read channel")
			}

		case apps.SubjectUserJoinedTeam, apps.SubjectUserLeftTeam:
			if sub.TeamID != "" && !mm.User.HasPermissionToTeam(userID, sub.TeamID, model.PermissionViewTeam) {
				return errors.New("no permission to view team")
			}

		case apps.SubjectBotJoinedChannelDeprecated,
			apps.SubjectBotLeftChannelDeprecated,
			apps.SubjectBotJoinedTeamDeprecated,
			apps.SubjectBotLeftTeamDeprecated:
			app, err := a.apps.Get(r.SourceAppID())
			if err != nil {
				return errors.Wrapf(err, "failed to get app %s to validate subscription to %s", r.SourceAppID(), sub.Subject)
			}
			if r.ActingUserID() != app.BotUserID {
				return errors.Errorf("%s can only be subscribed to by the app's bot", sub.Subject)
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
