// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

type UserLeftTeamNotification struct {
	SubscriptionID SubscriptionID
	Subject        SubscriptionSubject
	UserID         string
	TeamID         string
	Expanded       *Expanded
}

// OnUserLeftTeam sends a change notification when a new user has
// left a team.
func (p *proxy) OnUserLeftTeam(ctx *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	subs, err := p.Subscriptions.GetSubscriptionsForChannelOrTeam(SubjectUserLeftTeam, tm.TeamId)
	if err != nil {
		// p.Logger.Debugf("OnUserHasLeftTeam: failed to get subscriptions: %s %s: ",
		// 	SubjectUserLeftTeam, tm.TeamId, err)
		return
	}
	if len(subs) == 0 {
		return
	}

	expander := NewExpander(p.mm, p.Configurator)

	for _, s := range subs {
		expanded, err := expander.Expand(s.Expand, actingUser.Id, tm.UserId, tm.TeamId)
		if err != nil {
			// <><> TODO log
			return
		}

		msg := UserLeftTeamNotification{
			UserID:   tm.UserId,
			TeamID:   tm.TeamId,
			Expanded: expanded,
		}

		go p.SendChangeNotification(s, msg)
	}
}
