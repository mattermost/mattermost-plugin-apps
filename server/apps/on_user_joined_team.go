// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

type UserJoinedTeamNotification struct {
	SubscriptionID SubscriptionID
	Subject        SubscriptionSubject
	UserID         string
	TeamID         string
	Expanded       *Expanded
}

// OnUserJoinedTeam sends a change notification when a new user has joined a team
func (s *Service) OnUserJoinedTeam(ctx *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	subs, err := s.Subscriptions.GetChannelOrTeamSubs(SubjectUserJoinedTeam, tm.TeamId)
	if err != nil {
		// p.Logger.Debugf("OnUserHasJoinedTeam: failed to get subscriptions: %s %s: ",
		// 	SubjectUserJoinedTeam, tm.TeamId, err)
		return
	}

	for _, sub := range subs {
		expanded, err := s.Expander.Expand(sub.Expand, actingUser.Id, tm.UserId, tm.TeamId)
		if err != nil {
			// <><> TODO log
			return
		}

		msg := UserJoinedTeamNotification{
			UserID:   tm.UserId,
			TeamID:   tm.TeamId,
			Expanded: expanded,
		}

		go s.PostChangeNotification(*sub, msg)
	}
}
