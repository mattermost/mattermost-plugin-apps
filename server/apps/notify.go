// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

type Notification struct {
	Subject store.Subject
	Context *Context
}

func (s *service) NotifySubscribedApps(subj store.Subject, cc *Context) error {
	subs, err := s.Store.GetSubs(subj, cc.TeamID, cc.ChannelID)
	if err != nil {
		return err
	}

	expander := s.newExpander(cc)
	for _, sub := range subs {
		req := Notification{
			Subject: subj,
			Context: &Context{},
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
