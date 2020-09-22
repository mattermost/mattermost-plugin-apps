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

func (s *service) OnUserJoinedChannel(ctx *plugin.Context, cm *model.ChannelMember,
	actingUser *model.User) {
	subs, err := s.Subscriptions.GetSubscriptionsForChannel(SubjectUserJoinedChannel, cm.ChannelId)
	if err != nil {
		// p.Logger.Debugf("OnUserHasJoinedChannel: failed to get subscriptions: %s %s: ",
		// 	SubjectUserJoinedChannel, channelMember.ChannelId, err)
		return
	}
	if len(subs) == 0 {
		return
	}

	expander := NewExpander(s.Mattermost, s.Config)

	for _, sub := range subs {
		expanded, err := expander.Expand(sub.Expand, actingUser.Id, cm.UserId, cm.ChannelId)
		if err != nil {
			// <><> TODO log
			return
		}

		msg := UserJoinedChannelNotification{
			UserID:    cm.UserId,
			ChannelID: cm.ChannelId,
			Expanded:  expanded,
		}

		go s.Proxy.SendChangeNotification(sub, msg)
	}
}
