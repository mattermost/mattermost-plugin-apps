// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

type ExpandLevel string

const (
	ExpandDefault = ExpandLevel("")
	ExpandNone    = ExpandLevel("none")
	ExpandAll     = ExpandLevel("all")
	ExpandSummary = ExpandLevel("summary")
)

type Expand struct {
	App        ExpandLevel `json:"app,omitempty"`
	ActingUser ExpandLevel `json:"acting_user,omitempty"`

	// ActingUserAccessToken instruct the proxy to include OAuth2 access token
	// in the request. If the token is not available or is invalid, the user is
	// directed to the OAuth2 flow, and the Call is executed upon completion.
	ActingUserAccessToken ExpandLevel `json:"acting_user_access_token,omitempty"`

	// AdminAccessToken instructs the proxy to include an admin access token.
	AdminAccessToken ExpandLevel `json:"admin_access_token,omitempty"`

	Channel    ExpandLevel `json:"channel,omitempty"`
	Mentioned  ExpandLevel `json:"mentioned,omitempty"`
	ParentPost ExpandLevel `json:"parent_post,omitempty"`
	Post       ExpandLevel `json:"post,omitempty"`
	RootPost   ExpandLevel `json:"root_post,omitempty"`
	Team       ExpandLevel `json:"team,omitempty"`
	User       ExpandLevel `json:"user,omitempty"`
}
