// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

type ChannelCreatedNotification struct {
	SubscriptionID SubscriptionID
	Subject        SubscriptionSubject
	ChannelID      string
	TeamID         string
	Expanded       *Expanded
}

// OnChannelHasBeenCreated sends a change notification when a new channel has
// been created
func (p *proxy) OnChannelHasBeenCreated(ctx *plugin.Context, channel *model.Channel) {
	subs, err := p.Subscriptions.GetSubscriptionsForChannelOrTeam(SubjectChannelCreated, "")
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
		expanded, err := expander.Expand(s.Expand, "", "", "")
		if err != nil {
			// <><> TODO log
			return
		}

		msg := ChannelCreatedNotification{
			ChannelID: channel.Id,
			TeamID:    channel.TeamId,
			Expanded:  expanded,
		}

		go p.SendChangeNotification(s, msg)
	}
}
