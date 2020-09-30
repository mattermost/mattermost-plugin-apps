// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

type Hooks interface {
	OnUserHasBeenCreated(ctx *plugin.Context, user *model.User)
	OnUserJoinedChannel(pluginContext *plugin.Context, channelMember *model.ChannelMember, actor *model.User)
	OnUserLeftChannel(ctx *plugin.Context, cm *model.ChannelMember, actingUser *model.User)
	OnUserJoinedTeam(ctx *plugin.Context, tm *model.TeamMember, actingUser *model.User)
	OnUserLeftTeam(ctx *plugin.Context, tm *model.TeamMember, actingUser *model.User)
	SendNotifications(subs []*Subscription, cm *model.ChannelMember, actingUser *model.User, channel *model.Channel, post *model.Post, subject SubscriptionSubject)
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
	subs, err := s.Subscriptions.GetChannelOrTeamSubs(SubjectPostCreated, "")
	if err != nil {
		// p.Logger.Debugf("OnPostHasBeenCreated: failed to get subscriptions: %s %s: ",
		// 	SubjectPostCreated, user.UserId, err)
		return
	}
	s.SendNotifications(subs, nil, nil, nil, post, SubjectPostCreated)

}

// OnChannelHasBeenCreated sends a notification when a new channel has been created
func (s *Service) OnChannelHasBeenCreated(ctx *plugin.Context, channel *model.Channel) {
	subs, err := s.Subscriptions.GetChannelOrTeamSubs(SubjectChannelCreated, "")
	if err != nil {
		// p.Logger.Debugf("OnUserHasBeenCreated: failed to get subscriptions: %s %s: ",
		// 	SubjectUserCreated, user.UserId, err)
		return
	}
	s.SendNotifications(subs, nil, nil, channel, nil, SubjectChannelCreated)
}

// OnUserHasBeenCreated sends a notification when a new user has been created
func (s *Service) OnUserHasBeenCreated(ctx *plugin.Context, user *model.User) {
	subs, err := s.Subscriptions.GetChannelOrTeamSubs(SubjectUserCreated, "")
	if err != nil {
		// p.Logger.Debugf("OnUserHasBeenCreated: failed to get subscriptions: %s %s: ",
		// 	SubjectUserCreated, user.UserId, err)
		return
	}
	s.SendNotifications(subs, nil, user, nil, nil, SubjectUserCreated)
}

// OnUserJoinedChannel sends a notification when a new user has joined a channel
func (s *Service) OnUserJoinedChannel(ctx *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	subs, err := s.Subscriptions.GetChannelOrTeamSubs(SubjectUserJoinedChannel, cm.ChannelId)
	if err != nil {
		// p.Logger.Debugf("OnUserHasJoinedChannel: failed to get subscriptions: %s %s: ",
		// 	SubjectUserJoinedChannel, channelMember.ChannelId, err)
		return
	}
	s.SendNotifications(subs, cm, actingUser, nil, nil, SubjectUserJoinedChannel)
}

// OnUserLeftChannel sends a notification when a new user has left a channel
func (s *Service) OnUserLeftChannel(ctx *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	subs, err := s.Subscriptions.GetChannelOrTeamSubs(SubjectUserLeftChannel, cm.ChannelId)
	if err != nil {
		// p.Logger.Debugf("OnUserHasLeftChannel: failed to get subscriptions: %s %s: ",
		// 	SubjectUserLeftChannel, channelMember.ChannelId, err)
		return
	}
	s.SendNotifications(subs, cm, actingUser, nil, nil, SubjectUserLeftChannel)
}

// OnUserJoinedTeam sends a notification when a new user has joined a team
func (s *Service) OnUserJoinedTeam(ctx *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	subs, err := s.Subscriptions.GetChannelOrTeamSubs(SubjectUserJoinedTeam, tm.TeamId)
	if err != nil {
		// p.Logger.Debugf("OnUserHasJoinedTeam: failed to get subscriptions: %s %s: ",
		// 	SubjectUserJoinedTeam, tm.TeamId, err)
		return
	}
	s.SendNotifications(subs, nil, actingUser, nil, nil, SubjectUserJoinedTeam)
}

// OnUserLeftTeam sends a notification when a new user has left a team
func (s *Service) OnUserLeftTeam(ctx *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	subs, err := s.Subscriptions.GetChannelOrTeamSubs(SubjectUserLeftTeam, tm.TeamId)
	if err != nil {
		// p.Logger.Debugf("OnUserHasLeftTeam: failed to get subscriptions: %s %s: ",
		// 	SubjectUserLeftTeam, tm.TeamId, err)
		return
	}
	s.SendNotifications(subs, nil, actingUser, nil, nil, SubjectUserLeftTeam)
}

// SendNotifications sends a POST change notifiation for a set of subscriptions
func (s *Service) SendNotifications(subs []*Subscription, cm *model.ChannelMember, actingUser *model.User, channel *model.Channel, post *model.Post, subject SubscriptionSubject) {
	expander := NewExpander(s.Mattermost, s.Configurator)
	// TODO rectify the case where IDs exist from multiple function param inputs
	msg := &SubscriptionNotification{
		Subject:   subject,
		ChannelID: cm.ChannelId,
		ParentID:  post.ParentId,
		PostID:    post.Id,
		RootID:    post.RootId,
		TeamID:    channel.TeamId,
		UserID:    actingUser.Id,
	}

	for _, sub := range subs {
		expanded, err := expander.Expand(sub.Expand, actingUser.Id, cm.UserId, cm.ChannelId)
		if err != nil {
			// <><> TODO log
			return
		}
		msg.SubscriptionID = sub.SubscriptionID
		msg.Expanded = expanded

		go s.PostChangeNotification(*sub, msg)
	}
}
