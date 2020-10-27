// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-server/v5/model"
)

func (s *service) Notify(cc *api.Context, subj api.Subject) error {
	subs, err := s.Store.GetSubs(subj, cc.TeamID, cc.ChannelID)
	if err != nil {
		return err
	}

	expander := s.newExpander(cc)
	for _, sub := range subs {
		req := api.Notification{
			Subject: subj,
			Context: &api.Context{},
		}
		req.Context, err = expander.Expand(sub.Expand)
		if err != nil {
			return err
		}

		// Always set the AppID for routing the request to the App
		req.Context.AppID = sub.AppID

		go func() {
			_ = s.Client.PostNotification(&req)
		}()
	}
	return nil
}

func NewChannelContext(ch *model.Channel) *api.Context {
	return &api.Context{
		UserID:    ch.CreatorId,
		ChannelID: ch.Id,
		TeamID:    ch.TeamId,
		ExpandedContext: api.ExpandedContext{
			Channel: ch,
		},
	}
}

func NewPostContext(p *model.Post) *api.Context {
	return &api.Context{
		UserID:     p.UserId,
		PostID:     p.Id,
		RootPostID: p.RootId,
		ChannelID:  p.ChannelId,
		ExpandedContext: api.ExpandedContext{
			Post: p,
		},
	}
}

func NewUserContext(user *model.User) *api.Context {
	return &api.Context{
		UserID: user.Id,
		ExpandedContext: api.ExpandedContext{
			User: user,
		},
	}
}

func NewTeamMemberContext(tm *model.TeamMember, actingUser *model.User) *api.Context {
	actingUserID := ""
	if actingUser != nil {
		actingUserID = actingUser.Id
	}
	return &api.Context{
		ActingUserID: actingUserID,
		UserID:       tm.UserId,
		TeamID:       tm.TeamId,
		ExpandedContext: api.ExpandedContext{
			ActingUser: actingUser,
		},
	}
}

func NewChannelMemberContext(cm *model.ChannelMember, actingUser *model.User) *api.Context {
	actingUserID := ""
	if actingUser != nil {
		actingUserID = actingUser.Id
	}
	return &api.Context{
		ActingUserID: actingUserID,
		UserID:       cm.UserId,
		ChannelID:    cm.ChannelId,
		ExpandedContext: api.ExpandedContext{
			ActingUser: actingUser,
		},
	}
}
