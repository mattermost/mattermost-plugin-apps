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
	TeamID         string
	Expanded       *Expanded
}

// OnUserJoinedTeam sends a change notification when a new user has
// joined a team.
func (p *proxy) OnUserHasBeenCreated(ctx *plugin.Context, user *model.User) {
	subs, err := p.Subscriptions.GetSubscriptionsForChannelOrTeam(SubjectUserCreated, "")
	if err != nil {
		// p.Logger.Debugf("OnUserHasBeenCreated: failed to get subscriptions: %s %s: ",
		// 	SubjectUserCreated, user.UserId, err)
		return
	}
	if len(subs) == 0 {
		return
	}

	expander := NewExpander(p.mm, p.Configurator)

	for _, s := range subs {
		expanded, err := expander.Expand(s.Expand, user.Id, "", "")
		if err != nil {
			// <><> TODO log
			return
		}

		msg := UserCreatedNotification{
			// UserID:   tm.UserId,
			// TeamID:   tm.TeamId,
			Expanded: expanded,
		}

		go p.SendChangeNotification(s, msg)
	}
}
