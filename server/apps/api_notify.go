// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-server/v5/model"
)

type SubscriptionNotification struct {
	SubscriptionID SubscriptionID
	Subject        constants.SubscriptionSubject
	ChannelID      string
	ParentID       string
	PostID         string
	RootID         string
	TeamID         string
	UserID         string
	Expanded       *Expanded
}

// Notify sends a POST change notification for a set of subscriptions
func (s *Service) Notify(subject constants.SubscriptionSubject, tm *model.TeamMember,
	cm *model.ChannelMember, actingUser *model.User, channel *model.Channel, post *model.Post) error {
	actingUserID := ""
	channelOrTeamID := ""

	msg := &SubscriptionNotification{
		Subject: subject,
	}

	if actingUser != nil && actingUser.Id != "" {
		actingUserID = actingUser.Id
	}
	if channel != nil {
		msg.ChannelID = channel.Id
	}
	if cm != nil {
		msg.ChannelID = cm.ChannelId
		channelOrTeamID = cm.ChannelId
	}
	if tm != nil {
		msg.TeamID = tm.TeamId
		msg.UserID = tm.UserId
		channelOrTeamID = tm.TeamId
	}
	if post != nil {
		msg.UserID = post.UserId
		msg.PostID = post.Id
		msg.ParentID = post.ParentId
		msg.RootID = post.RootId
	}

	// get subscriptions for the given subject
	subs, err := s.Subscriptions.GetChannelOrTeamSubs(subject, channelOrTeamID)
	if err != nil {
		// p.Logger.Debugf("OnPostHasBeenCreated: failed to get subscriptions: %s %s: ",
		// 	SubjectPostCreated, user.UserId, err)
		return nil
	}

	expander := NewExpander(s.Mattermost, s.Configurator)
	for _, sub := range subs {
		// only expand if sub requests it
		if sub.Expand != nil {
			expanded, err := expander.Expand(sub.Expand, actingUserID, msg.UserID, msg.ChannelID, msg.TeamID)
			if err != nil {
				// <><> TODO log
				return nil
			}
			msg.Expanded = expanded
		}
		go s.PostChangeNotification(*sub, msg)
	}
	return nil
}
