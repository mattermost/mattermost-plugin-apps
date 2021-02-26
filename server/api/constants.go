// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

// Internal configuration apps.of mattermost-plugin-apps
const (
	Repository     = "mattermost-plugin-apps"
	CommandTrigger = "apps"

	BotUsername    = "appsbot"
	BotDisplayName = "Mattermost Apps"
	BotDescription = "Mattermost Apps Registry and API proxy."

	// TODO replace Interactive Dialogs with Modal, eliminate the need for
	// /dialog endpoints.
	InteractiveDialogPath = "/dialog"

	// Top-level path(s) for HTTP example apps.
	HelloHTTPPath = "/example/hello"

	// Top-level path for the REST APIs exposed by the plugin itself.
	APIPath = "/api/v1"

	// Top-level path for the Apps namespaces, followed by /AppID/subpath.
	AppsPath = "/apps"

	// OAuth2 sub-paths.
	PathOAuth2         = "/oauth2"          // convention for Mattermost Apps, comes from OAuther
	PathOAuth2Complete = "/oauth2/complete" // convention for Mattermost Apps, comes from OAuther

	// Marketplace sub-paths.
	PathMarketplace = "/marketplace"

	// Other sub-paths.
	CallPath        = "/call"
	KVPath          = "/kv"
	SubscribePath   = "/subscribe"
	UnsubscribePath = "/unsubscribe"

	BindingsPath = "/bindings"

	WesocketEventRefreshBindings = "refresh_bindings"
)

const (
	PropTeamID    = "team_id"
	PropChannelID = "channel_id"
	PropPostID    = "post_id"
)
