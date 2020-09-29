// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

type UserCreatedNotification struct {
	SubscriptionID SubscriptionID
	Subject        SubscriptionSubject
	UserID         string
	Expanded       *Expanded
}

// OnUserHasBeenCreated sends a change notification when a new user has
// joined a team.
func (s *Service) OnUserHasBeenCreated(ctx *plugin.Context, user *model.User) {
	subs, err := s.Subscriptions.GetChannelOrTeamSubs(SubjectUserCreated, "")
	if err != nil {
		// p.Logger.Debugf("OnUserHasBeenCreated: failed to get subscriptions: %s %s: ",
		// 	SubjectUserCreated, user.UserId, err)
		return
	}

	for _, sub := range subs {
		expanded, err := s.Expander.Expand(sub.Expand, "", "", "")
		if err != nil {
			// <><> TODO log
			return
		}

		msg := UserCreatedNotification{
			UserID:   user.Id,
			Expanded: expanded,
		}

		go s.PostChangeNotification(*sub, msg)
	}
}
