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
func (p *proxy) OnUserHasBeenCreated(ctx *plugin.Context, user *model.User) {
	subs, err := p.Subscriptions.GetSubsForChannelOrTeam(SubjectUserCreated, "")
	if err != nil {
		// p.Logger.Debugf("OnUserHasBeenCreated: failed to get subscriptions: %s %s: ",
		// 	SubjectUserCreated, user.UserId, err)
		return
	}

	expander := NewExpander(p.mm, p.configurator)

	for _, s := range subs {
		expanded, err := expander.Expand(s.Expand, "", "", "")
		if err != nil {
			// <><> TODO log
			return
		}

		msg := UserCreatedNotification{
			UserID:   user.Id,
			Expanded: expanded,
		}

		go p.SendChangeNotification(s, msg)
	}
}
