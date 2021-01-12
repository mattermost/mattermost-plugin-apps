// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package impl

import (
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/pkg/errors"
)

func (s *service) filterContext(c *apps.Call) error {
	checkTeam := true
	if c.Context.ChannelID != "" {
		_, err := s.Mattermost.Channel.GetMember(c.Context.ChannelID, c.Context.ActingUserID)
		if err != nil {
			return errors.Wrap(err, "user is not a member of channel specified in context")
		}

		ch, err := s.Mattermost.Channel.Get(c.Context.ChannelID)
		if err != nil {
			return errors.Wrap(err, "failed to fetch channel specified in context")
		}

		if ch.TeamId != "" {
			checkTeam = false
			c.Context.TeamID = ch.TeamId
		}
	}

	if checkTeam && c.Context.TeamID != "" {
		_, err := s.Mattermost.Team.GetMember(c.Context.TeamID, c.Context.ActingUserID)
		if err != nil {
			return errors.Wrap(err, "user is not a member of team specified in context")
		}
	}

	return nil
}

func (s *service) Call(c *apps.Call) (*apps.CallResponse, error) {
	err := s.filterContext(c)
	if err != nil {
		return nil, err
	}

	cc, err := s.newExpander(c.Context).Expand(c.Expand)
	if err != nil {
		return nil, err
	}

	clone := *c
	clone.Context = cc
	return s.Client.PostCall(&clone)
}

func (s *service) Notify(cc *apps.Context, subj apps.Subject) error {
	subs, err := s.Store.GetSubs(subj, cc.TeamID, cc.ChannelID)
	if err != nil {
		return err
	}

	expander := s.newExpander(cc)
	for _, sub := range subs {
		req := apps.Notification{
			Subject: subj,
			Context: &apps.Context{},
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
