// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

// Internal configuration apps.of mattermost-plugin-apps
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
	PropAppID              = "app_id"
	PropTeamID             = "team_id"
	PropChannelID          = "channel_id"
	PropActingUserID       = "acting_user_id"
	PropPostID             = "post_id"
	PropBotAccessToken     = "bot_access_token"
	PropOAuth2ClientSecret = "oauth2_client_secret" // nolint:gosec
	PropAppBindings        = "app_bindings"
)
