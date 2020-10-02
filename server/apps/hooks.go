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
	OnChannelHasBeenCreated(ctx *plugin.Context, channel *model.Channel)
	OnPostHasBeenCreated(ctx *plugin.Context, post *model.Post)
	SendNotifications(
		subject SubscriptionSubject,
		subs []*Subscription,
		tm *model.TeamMember,
		cm *model.ChannelMember,
		actingUser *model.User,
		channel *model.Channel,
		post *model.Post,
	)
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
	s.SendNotifications(SubjectPostCreated, subs, nil, nil, nil, nil, post)
}

// OnChannelHasBeenCreated sends a notification when a new channel has been created
func (s *Service) OnChannelHasBeenCreated(ctx *plugin.Context, channel *model.Channel) {
	subs, err := s.Subscriptions.GetChannelOrTeamSubs(SubjectChannelCreated, "")
	if err != nil {
		// p.Logger.Debugf("OnUserHasBeenCreated: failed to get subscriptions: %s %s: ",
		// 	SubjectUserCreated, user.UserId, err)
		return
	}
	s.SendNotifications(SubjectChannelCreated, subs, nil, nil, nil, channel, nil)
}

// OnUserHasBeenCreated sends a notification when a new user has been created
func (s *Service) OnUserHasBeenCreated(ctx *plugin.Context, user *model.User) {
	subs, err := s.Subscriptions.GetChannelOrTeamSubs(SubjectUserCreated, "")
	if err != nil {
		// p.Logger.Debugf("OnUserHasBeenCreated: failed to get subscriptions: %s %s: ",
		// 	SubjectUserCreated, user.UserId, err)
		return
	}
	s.SendNotifications(SubjectUserCreated, subs, nil, nil, user, nil, nil)
}

// OnUserJoinedChannel sends a notification when a new user has joined a channel
func (s *Service) OnUserJoinedChannel(ctx *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	actingUserD, _ := json.MarshalIndent(actingUser, "", "    ")
	fmt.Printf("actingUser = %+v\n", string(actingUserD))
	subs, err := s.Subscriptions.GetChannelOrTeamSubs(SubjectUserJoinedChannel, cm.ChannelId)
	if err != nil {
		// p.Logger.Debugf("OnUserHasJoinedChannel: failed to get subscriptions: %s %s: ",
		// 	SubjectUserJoinedChannel, channelMember.ChannelId, err)
		return
	}
	s.SendNotifications(SubjectUserJoinedChannel, subs, nil, cm, actingUser, nil, nil)
}

// OnUserLeftChannel sends a notification when a new user has left a channel
func (s *Service) OnUserLeftChannel(ctx *plugin.Context, cm *model.ChannelMember, actingUser *model.User) {
	subs, err := s.Subscriptions.GetChannelOrTeamSubs(SubjectUserLeftChannel, cm.ChannelId)
	if err != nil {
		// p.Logger.Debugf("OnUserHasLeftChannel: failed to get subscriptions: %s %s: ",
		// 	SubjectUserLeftChannel, channelMember.ChannelId, err)
		return
	}
	s.SendNotifications(SubjectUserLeftChannel, subs, nil, cm, actingUser, nil, nil)
}

// OnUserJoinedTeam sends a notification when a new user has joined a team
func (s *Service) OnUserJoinedTeam(ctx *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	subs, err := s.Subscriptions.GetChannelOrTeamSubs(SubjectUserJoinedTeam, tm.TeamId)
	if err != nil {
		// p.Logger.Debugf("OnUserHasJoinedTeam: failed to get subscriptions: %s %s: ",
		// 	SubjectUserJoinedTeam, tm.TeamId, err)
		return
	}
	s.SendNotifications(SubjectUserJoinedTeam, subs, tm, nil, actingUser, nil, nil)
}

// OnUserLeftTeam sends a notification when a new user has left a team
func (s *Service) OnUserLeftTeam(ctx *plugin.Context, tm *model.TeamMember, actingUser *model.User) {
	subs, err := s.Subscriptions.GetChannelOrTeamSubs(SubjectUserLeftTeam, tm.TeamId)
	if err != nil {
		// p.Logger.Debugf("OnUserHasLeftTeam: failed to get subscriptions: %s %s: ",
		// 	SubjectUserLeftTeam, tm.TeamId, err)
		return
	}
	s.SendNotifications(SubjectUserLeftTeam, subs, tm, nil, actingUser, nil, nil)
}

// SendNotifications sends a POST change notifiation for a set of subscriptions
func (s *Service) SendNotifications(subject SubscriptionSubject, subs []*Subscription, tm *model.TeamMember, cm *model.ChannelMember, actingUser *model.User, channel *model.Channel, post *model.Post) {

	fmt.Printf("Subject = %+v\n", subject)

	// TODO rectify the case where IDs exist from multiple function param inputs
	actingUserID := ""
	if actingUser != nil && actingUser.Id != "" {
		actingUserID = actingUser.Id
	}

	cmUserID := ""
	cmChannelID := ""
	if cm != nil {
		if cm.UserId != "" {
			cmUserID = cm.UserId
		}
		if cm.ChannelId != "" {
			cmChannelID = cm.ChannelId
		}
	}
	msg := &SubscriptionNotification{
		Subject:   subject,
		ChannelID: cmChannelID,
		ParentID:  post.ParentId,
		PostID:    post.Id,
		RootID:    post.RootId,
		// TeamID:    channel.TeamId,
		UserID: actingUserID,
	}

	expander := NewExpander(s.Mattermost, s.Configurator)
	for _, sub := range subs {
		subD, _ := json.MarshalIndent(sub, "", "    ")
		fmt.Printf("sub = %+v\n", string(subD))

		// only expand if sub requests it
		if sub.Expand != nil {
			expanded, err := expander.Expand(sub.Expand, actingUserID, cmUserID, cmChannelID)
			if err != nil {
				// <><> TODO log
				return
			}
			msg.Expanded = expanded
		}
		msg.SubscriptionID = sub.SubscriptionID
		msg.Expanded = expanded

		go s.PostChangeNotification(*sub, msg)
	}
}
