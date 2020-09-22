// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package appmodel

import "github.com/mattermost/mattermost-server/v5/model"

type Expander interface {
	Expand(expand *Expand, actingUserID, userID, channelID string) (*Expanded, error)
}

type ExpandEntity string

const (
	ExpandActingUser = ExpandEntity("ActingUser")
	ExpandUser       = ExpandEntity("User")
	ExpandChannel    = ExpandEntity("Channel")
	ExpandConfig     = ExpandEntity("Config")
)

type ExpandLevel string

const (
	ExpandAll     = ExpandLevel("All")
	ExpandSummary = ExpandLevel("Summary")
)

type Expand struct {
	ActingUser ExpandLevel
	Channel    ExpandLevel
	Config     bool
	User       ExpandLevel
	Post       ExpandLevel
	ParentPost ExpandLevel
	RootPost   ExpandLevel
	Team       ExpandLevel
	Mentioned  ExpandLevel
}

type Expanded struct {
	ActingUser *model.User
	Channel    *model.Channel
	Config     *MattermostConfig
	User       *model.User
	Post       *model.Post
	ParentPost *model.Post
	RootPost   *model.Post
	Team       *model.Team
	Mentioned  []*model.User
}
