// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"encoding/json"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

type Hooks interface {
	OnUserHasBeenCreated(ctx *plugin.Context, user *model.User)
	OnUserJoinedChannel(pluginContext *plugin.Context, channelMember *model.ChannelMember, actor *model.User)
	OnUserLeftChannel(ctx *plugin.Context, cm *model.ChannelMember, actingUser *model.User)
	OnUserJoinedTeam(ctx *plugin.Context, tm *model.TeamMember, actingUser *model.User)
	OnUserLeftTeam(ctx *plugin.Context, tm *model.TeamMember, actingUser *model.User)
	OnChannelHasBeenCreated(ctx *plugin.Context, channel *model.Channel)
	OnPostHasBeenCreated(ctx *plugin.Context, post *model.Post)
	SendNotifications(
		subject SubscriptionSubject,
		tm *model.TeamMember,
		cm *model.ChannelMember,
		actingUser *model.User,
		channel *model.Channel,
		post *model.Post,
	) error
}

type SubscriptionNotification struct {
	SubscriptionID SubscriptionID
	Subject        SubscriptionSubject
	ChannelID      string
	ParentID       string
	PostID         string
	RootID         string
	TeamID         string
	UserID         string
	Expanded       *Expanded
}

// OnPostHasBeenCreated sends a notification when a new post has been created
func (s *Service) OnPostHasBeenCreated(ctx *plugin.Context, post *model.Post) {
	err := s.SendNotifications(SubjectPostCreated, nil, nil, nil, nil, post)
	if err != nil {
		// p.Logger.Debugf("OnPostHasBeenCreated: failed to get subscriptions: %s %s: ",
		// 	SubjectPostCreated, user.UserId, err)
		return
	}

}

// OnChannelHasBeenCreated sends a notification when a new channel has been created
func (s *Service) OnChannelHasBeenCreated(ctx *plugin.Context, channel *model.Channel) {
	s.SendNotifications(SubjectChannelCreated, nil, nil, nil, channel, nil)
}

// OnUserHasBeenCreated sends a notification when a new user has been created
func (s *Service) OnUserHasBeenCreated(ctx *plugin.Context, user *model.User) {
	s.SendNotifications(SubjectUserCreated, nil, nil, user, nil, nil)
}

// OnUserJoinedChannel sends a notification when a new user has joined a channel
func (s *Service) OnUserJoinedChannel(ctx *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	s.SendNotifications(SubjectUserJoinedChannel, nil, cm, actingUser, nil, nil)
}

// OnUserLeftChannel sends a notification when a new user has left a channel
func (s *Service) OnUserLeftChannel(ctx *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	s.SendNotifications(SubjectUserLeftChannel, nil, cm, actingUser, nil, nil)
}

// OnUserJoinedTeam sends a notification when a new user has joined a team
func (s *Service) OnUserJoinedTeam(ctx *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	s.SendNotifications(SubjectUserJoinedTeam, tm, nil, actingUser, nil, nil)
}

// OnUserLeftTeam sends a notification when a new user has left a team
func (s *Service) OnUserLeftTeam(ctx *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	s.SendNotifications(SubjectUserLeftTeam, tm, nil, actingUser, nil, nil)
}

// SendNotifications sends a POST change notifiation for a set of subscriptions
func (s *Service) SendNotifications(subject SubscriptionSubject, tm *model.TeamMember, cm *model.ChannelMember, actingUser *model.User, channel *model.Channel, post *model.Post) error {

	actingUserID := ""
	cmChannelID := ""
	channelOrTeamID := ""

	msg := &SubscriptionNotification{
		Subject: subject,
		UserID:  actingUserID,
	}

	if actingUser != nil && actingUser.Id != "" {
		actingUserID = actingUser.Id
	}
	if cm != nil {
		msg.ChannelID = cm.ChannelId
		channelOrTeamID = cmChannelID
	}
	if tm != nil {
		msg.TeamID = tm.TeamId
		channelOrTeamID = tm.TeamId
	}
	if post != nil {
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
			expanded, err := expander.Expand(sub.Expand, actingUserID, msg.UserID, cmChannelID)
			if err != nil {
				// <><> TODO log
				return nil
			}
			msg.Expanded = expanded
		}

		msgD, _ := json.MarshalIndent(msg, "", "    ")
		go s.PostChangeNotification(*sub, msg)
	}
	return nil
}
