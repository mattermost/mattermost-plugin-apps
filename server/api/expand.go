// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import "github.com/mattermost/mattermost-server/v5/model"

type ExpandLevel string

// <><> TODO update ExpandLevels in redux
const (
	ExpandDefault = ExpandLevel("")
	ExpandNone    = ExpandLevel("none")
	ExpandAll     = ExpandLevel("all")
	ExpandSummary = ExpandLevel("summary")
)

type Expand struct {
	// Expanded App contains
	App        ExpandLevel `json:"app"`
	ActingUser ExpandLevel `json:"acting_user"`

	// ActingUserAccessToken instruct the proxy to include OAuth2 access token
	// in the request. If the token is not available or is invalid, the user is
	// directed to the OAuth2 flow, and the Call is executed upon completion.
	ActingUserAccessToken ExpandLevel `json:"acting_user_access_token"`

	// AdminAccessToken instructs the proxy to include an admin access token.
	AdminAccessToken ExpandLevel `json:"admin_access_token"`

	Channel    ExpandLevel `json:"channel,omitempty"`
	Config     ExpandLevel `json:"config,omitempty"`
	Mentioned  ExpandLevel `json:"mentioned,omitempty"`
	ParentPost ExpandLevel `json:"parent_post,omitempty"`
	Post       ExpandLevel `json:"post,omitempty"`
	RootPost   ExpandLevel `json:"root_post,omitempty"`
	Team       ExpandLevel `json:"team,omitempty"`
	User       ExpandLevel `json:"user,omitempty"`
}

type ExpandedContext struct {
	BotAccessToken        string            `json:"bot_access_token,omitempty"`
	AdminAccessToken      string            `json:"admin_access_token,omitempty"`
	ActingUserAccessToken string            `json:"acting_user_access_token,omitempty"`
	ActingUser            *model.User       `json:"acting_user,omitempty"`
	App                   *App              `json:"app,omitempty"`
	Channel               *model.Channel    `json:"channel,omitempty"`
	Config                *MattermostConfig `json:"config,omitempty"`
	Mentioned             []*model.User     `json:"mentioned,omitempty"`
	Post                  *model.Post       `json:"post,omitempty"`
	RootPost              *model.Post       `json:"root_post,omitempty"`
	Team                  *model.Team       `json:"team,omitempty"`
	// TODO replace User with mentions
	User *model.User `json:"user,omitempty"`
}
