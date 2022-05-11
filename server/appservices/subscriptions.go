// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package appservices

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

type PermissionChecker interface {
	HasPermissionTo(userID string, permission *model.Permission) bool
	HasPermissionToChannel(userID, channelID string, permission *model.Permission) bool
	HasPermissionToTeam(userID, teamID string, permission *model.Permission) bool
}

func CheckSubscriptionPermission(checker PermissionChecker, sub apps.Subscription, channelID, teamID string) error {
	userID := sub.UserID

	switch sub.Subject {
	case apps.SubjectUserCreated:
		if !checker.HasPermissionTo(userID, model.PermissionViewMembers) {
			return errors.New("no permission to read user")
		}
	case apps.SubjectUserJoinedChannel,
		apps.SubjectUserLeftChannel,
		apps.SubjectPostCreated:
		if !checker.HasPermissionToChannel(userID, sub.ChannelID, model.PermissionReadChannel) {
			return errors.New("no permission to read channel")
		}
	case apps.SubjectBotJoinedChannel,
		apps.SubjectBotLeftChannel,
		apps.SubjectBotMentioned:
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

func (a *AppServices) Subscribe(_ *incoming.Request, sub apps.Subscription) error {
	if err := sub.Validate(); err != nil {
		return err
	}

	if err := CheckSubscriptionPermission(&a.conf.MattermostAPI().User, sub, sub.ChannelID, sub.TeamID); err != nil {
		return err
	}

	return a.store.Subscription.Save(sub)
}

func (a *AppServices) GetSubscriptions(r *incoming.Request) ([]apps.Subscription, error) {
	return a.store.Subscription.ListByUserID(r.SourceAppID(), r.ActingUserID())
}

func (a *AppServices) Unsubscribe(r *incoming.Request, sub apps.Subscription) error {
	if err := sub.Validate(); err != nil {
		return err
	}

	return a.store.Subscription.Delete(sub)
}
