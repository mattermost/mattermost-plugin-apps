// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package impl

import (
	"encoding/json"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-server/v5/model"
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

	resp, err := s.Client.PostCall(&clone)
	if err != nil {
		return nil, err
	}
	s.updatePostIfNeeded(c, resp)

	return resp, nil
}

func (s *service) updatePostIfNeeded(c *apps.Call, resp *apps.CallResponse) {
	if apps.LocationInPost.In(c.Context.Location) && resp.Type == apps.CallResponseTypeUpdateEmbedded {
		if resp.Data[apps.EmbeddedResponseDataPost] != nil {
			if updatedPost, parseErr := postFromInterface(resp.Data[apps.EmbeddedResponseDataPost]); parseErr == nil {
				updatedPost.Id = c.Context.PostID
				// TODO More checks on the post to use for Update
				err := s.Mattermost.Post.UpdatePost(updatedPost)
				s.Mattermost.Log.Debug("error updating", "error", err)
				// TODO Log error?
			}
		}
	}
}

func postFromInterface(v interface{}) (*model.Post, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var post model.Post
	err = json.Unmarshal(b, &post)
	if err != nil {
		return nil, err
	}

	return &post, nil
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
