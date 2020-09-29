// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

type UserLeftChannelNotification struct {
	SubscriptionID SubscriptionID
	Subject        SubscriptionSubject
	UserID         string
	ChannelID      string
	Expanded       *Expanded
}

// OnUserLeftChannel sends a change notification when a new user has left a channel
func (s *Service) OnUserLeftChannel(ctx *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	subs, err := s.Subscriptions.GetChannelOrTeamSubs(SubjectUserLeftChannel, cm.ChannelId)
	if err != nil {
		// p.Logger.Debugf("OnUserHasLeftChannel: failed to get subscriptions: %s %s: ",
		// 	SubjectUserLeftChannel, channelMember.ChannelId, err)
		return
	}

	for _, sub := range subs {
		expanded, err := s.Expander.Expand(sub.Expand, actingUser.Id, cm.UserId, cm.ChannelId)
		if err != nil {
			// <><> TODO log
			return
		}

		msg := UserLeftChannelNotification{
			UserID:    actingUser.Id,
			ChannelID: cm.ChannelId,
			Expanded:  expanded,
		}

		go s.PostChangeNotification(*sub, msg)
	}
}
