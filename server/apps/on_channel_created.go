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

// OnChannelHasBeenCreated sends a change notification when a new channel has been created
func (s *Service) OnChannelHasBeenCreated(ctx *plugin.Context, channel *model.Channel) {
	subs, err := s.Subscriptions.GetChannelOrTeamSubs(SubjectChannelCreated, "")
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

		msg := ChannelCreatedNotification{
			ChannelID: channel.Id,
			TeamID:    channel.TeamId,
			Expanded:  expanded,
		}

		go s.PostChangeNotification(*sub, msg)
	}
}
