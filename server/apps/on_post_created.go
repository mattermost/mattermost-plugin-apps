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
func (p *proxy) OnPostHasBeenCreated(ctx *plugin.Context, post *model.Post) {
	subs, err := p.Subscriptions.GetSubscriptionsForChannelOrTeam(SubjectPostCreated, "")
	if err != nil {
		// p.Logger.Debugf("OnPostHasBeenCreated: failed to get subscriptions: %s %s: ",
		// 	SubjectPostCreated, user.UserId, err)
		return
	}
	if len(subs) == 0 {
		return
	}

	expander := NewExpander(p.mm, p.configurator)

	for _, s := range subs {
		expanded, err := expander.Expand(s.Expand, "", "", "")
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

		go p.SendChangeNotification(s, msg)
	}
}
