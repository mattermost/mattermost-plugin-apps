// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/mattermost/mattermost-server/v6/model"
)

// Context is included in CallRequest and provides App with information about
// Mattermost environment (configuration, authentication), and the context of
// the user agent (current channel, etc.)
//
// To help reduce the need to go back to Mattermost REST API, ExpandedContext
// can be included by adding a corresponding Expand attribute to the originating
// Call.
type Context struct {
	// Subject is a subject of notification, if the call originated from a
	// subscription.
	Subject Subject `json:"subject,omitempty"`

	// BotUserID of the App.
	BotUserID string `json:"bot_user_id"`

	// ActingUserID is primarily (or exclusively?) for calls originating from
	// user submissions.
	ActingUserID string `json:"acting_user_id,omitempty"`

	// UserID indicates the subject of the command. Once Mentions is
	// implemented, it may be replaced by Mentions.
	UserID string `json:"user_id,omitempty"`

	// Top-level Mattermost site URL to use for REST API calls.
	MattermostSiteURL string `json:"mattermost_site_url"`

	// App's path on the Mattermost instance (appendable to MattermostSiteURL).
	AppPath string `json:"app_path"`

	// Data accepted from the user agent
	UserAgentContext

	// More data as requested by call.Expand
	ExpandedContext
}

// UserAgentContext is a subset of fields from Context that are accepted from the user agent
// The values are vetted, and all fields present in the provided Context that are not in UserAgentContext are discarded when the Call comes from an acting user.
type UserAgentContext struct {
	// The optional IDs of Mattermost entities associated with the call: Team,
	// Channel, Post, RootPost.
	TeamID     string `json:"team_id"`
	ChannelID  string `json:"channel_id,omitempty"`
	PostID     string `json:"post_id,omitempty"`
	RootPostID string `json:"root_post_id,omitempty"`

	// AppID is used for handling CallRequest internally.
	AppID AppID `json:"app_id"`

	// Fully qualified original Location of the user action (if applicable),
	// e.g. "/command/helloworld/send" or "/channel_header/send".
	Location Location `json:"location,omitempty"`

	// UserAgent used to perform the call. It can be either "webapp" or "mobile".
	// Non user interactions like notifications will have this field empty.
	UserAgent string `json:"user_agent,omitempty"`
}

// ExpandedContext contains authentication, and Mattermost entity data, as
// indicated by the Expand attribute of the originating Call.
type ExpandedContext struct {
	//  BotAccessToken is always provided in expanded context
	BotAccessToken string `json:"bot_access_token,omitempty"`

	ActingUser            *model.User    `json:"acting_user,omitempty"`
	ActingUserAccessToken string         `json:"acting_user_access_token,omitempty"`
	AdminAccessToken      string         `json:"admin_access_token,omitempty"`
	OAuth2                OAuth2Context  `json:"oauth2,omitempty"`
	App                   *App           `json:"app,omitempty"`
	Channel               *model.Channel `json:"channel,omitempty"`
	Mentioned             []*model.User  `json:"mentioned,omitempty"`
	Post                  *model.Post    `json:"post,omitempty"`
	RootPost              *model.Post    `json:"root_post,omitempty"`
	Team                  *model.Team    `json:"team,omitempty"`
	Locale                string         `json:"locale,omitempty"`

	// TODO replace User with mentions
	User *model.User `json:"user,omitempty"`
}

type OAuth2Context struct {
	// Expanded with "oauth2_app". Config must be previously stored with
	// appclient.StoreOAuth2App
	OAuth2App
	ConnectURL  string `json:"connect_url,omitempty"`
	CompleteURL string `json:"complete_url,omitempty"`

	User interface{} `json:"user,omitempty"`
}
