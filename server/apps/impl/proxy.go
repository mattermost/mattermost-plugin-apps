// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package impl

import (
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
)

func (s *service) Call(c *apps.Call) (*apps.CallResponse, error) {
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
