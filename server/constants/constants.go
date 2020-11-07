// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package constants

// Internal configuration constants of mattermost-plugin-apps
const (
	Repository     = "mattermost-plugin-apps"
	CommandTrigger = "apps"

	BotUsername    = "appsbot"
	BotDisplayName = "Mattermost Apps"
	BotDescription = "Mattermost Apps Registry and API proxy."

	InteractiveDialogPath = "/dialog"
	HelloAppPath          = "/hello"
	APIPath               = "/api/v1"
	CallPath              = "/call"
	SubscribePath         = "/subscribe"
	BindingsPath          = "/bindings"
)

// Conventions for Apps paths, and field names
const (
	AppInstallPath  = "/install"
	AppBindingsPath = "/bindings"
)

const (
	AppID               = "app_id"
	TeamID              = "team_id"
	ChannelID           = "channel_id"
	ActingUserID        = "acting_user_id"
	PostID              = "post_id"
	BotAccessToken      = "bot_access_token"
	OAuth2ClientSecret  = "oauth2_client_secret" // nolint:gosec
	PostPropAppBindings = "app_bindings"
)
