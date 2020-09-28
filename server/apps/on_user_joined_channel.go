// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

type UserJoinedChannelNotification struct {
	SubscriptionID SubscriptionID
	Subject        SubscriptionSubject
	UserID         string
	ChannelID      string
	Expanded       *Expanded
}

// OnUserJoinedChannel sends a change notification when a new user has joined a channel
func (p *proxy) OnUserJoinedChannel(ctx *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	subs, err := p.Subscriptions.GetSubsForChannelOrTeam(SubjectUserJoinedChannel, cm.ChannelId)
	if err != nil {
		// p.Logger.Debugf("OnUserHasJoinedChannel: failed to get subscriptions: %s %s: ",
		// 	SubjectUserJoinedChannel, channelMember.ChannelId, err)
		return
	}

	expander := NewExpander(p.mm, p.configurator)

	for _, s := range subs {
		expanded, err := expander.Expand(s.Expand, actingUser.Id, cm.UserId, cm.ChannelId)
		if err != nil {
			// <><> TODO log
			return
		}

		msg := UserJoinedChannelNotification{
			UserID:    cm.UserId,
			ChannelID: cm.ChannelId,
			Expanded:  expanded,
		}

		go p.SendChangeNotification(s, msg)
	}
}
