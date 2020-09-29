// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

type PostCreatedNotification struct {
	SubscriptionID SubscriptionID
	Subject        SubscriptionSubject
	ChannelID      string
	ParentID       string
	PostID         string
	RootID         string
	UserID         string
	// TeamID         string
	Expanded *Expanded
}

// OnUserHasBeenCreated sends a change notification when a new user has
// joined a team.
func (s *Service) OnPostHasBeenCreated(ctx *plugin.Context, post *model.Post) {
	subs, err := s.Subscriptions.GetChannelOrTeamSubs(SubjectPostCreated, "")
	if err != nil {
		// p.Logger.Debugf("OnPostHasBeenCreated: failed to get subscriptions: %s %s: ",
		// 	SubjectPostCreated, user.UserId, err)
		return
	}

	for _, sub := range subs {
		expanded, err := s.Expander.Expand(sub.Expand, "", "", "")
		if err != nil {
			// <><> TODO log
			return
		}

		msg := PostCreatedNotification{
			PostID:    post.Id,
			UserID:    post.UserId,
			ChannelID: post.ChannelId,
			ParentID:  post.ParentId,
			RootID:    post.RootId,
			// TeamID:    post.TeamId, // doesn't exist in post
			Expanded: expanded,
		}

		go s.PostChangeNotification(*sub, msg)
	}
}
