// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package config

import "github.com/mattermost/mattermost-plugin-apps/apps"

const (
	MattermostSessionIDHeader = "Mattermost-Session-Id"
	MattermostPluginIDHeader  = "Mattermost-Plugin-Id"
	MattermostUserIDHeader    = "Mattermost-User-Id"
)

// Internal configuration apps.of mattermost-plugin-apps
const (
	Repository     = "mattermost-plugin-apps"
	CommandTrigger = "apps"
	ManifestsFile  = "manifests.json"

	BotUsername    = "appsbot"
	BotDisplayName = "Mattermost Apps"
	BotDescription = "Mattermost Apps Registry and API proxy."
)

const (
	// TODO replace Interactive Dialogs with Modal, eliminate the need for
	// /dialog endpoints.
	InteractiveDialogPath = "/dialog"

	// Top-level path(s) for HTTP example apps.
	HelloHTTPPath = "/example/hello"

	// Path to the Call API
	// <>/<> TODO: ticket migrate to gateway
	PathCall = "/call"

	// Top-level path for the Apps namespaces, followed by /{AppID}/...
	PathApps = "/apps"

	PingPath = "/ping"

	// OAuth2 App's HTTP endpoints in the {PluginURL}/apps/{AppID} space.
	PathMattermostOAuth2Connect  = "/oauth2/mattermost/connect"
	PathMattermostOAuth2Complete = "/oauth2/mattermost/complete"
	PathRemoteOAuth2Connect      = "/oauth2/remote/connect"
	PathRemoteOAuth2Complete     = "/oauth2/remote/complete"

	// Static assets are served from {PluginURL}/static/...
	PathStatic = "/" + apps.StaticFolder

	// Marketplace sub-paths.
	PathMarketplace = "/marketplace"

	WebSocketEventRefreshBindings = "refresh_bindings"
	WebSocketEventPluginEnabled   = "plugin_enabled"
	WebSocketEventPluginDisabled  = "plugin_disabled"
)

const (
	PropTeamID    = "team_id"
	PropChannelID = "channel_id"
	PropPostID    = "post_id"
	PropUserAgent = "user_agent_type"
)

// KV namespace
//
// Keys starting with a '.' are reserved for app-specific keys in the "hashkey"
// format. Hashkeys have the following format (see service_test.go#TestHashkey
// for examples):
//
//  - global prefix of ".X" where X is exactly 1 byte (2 bytes)
//  - bot user ID (26 bytes)
//  - app-specific prefix, limited to 2 non-space ASCII characters, right-filled
//   with ' ' to 2 bytes.
//  - app-specific key hash: 16 bytes, ascii85 (20 bytes)
//
// All other keys must start with an ASCII letter. '.' is usually used as the
// terminator since it is not used in the base64 representation.
const (
	// KVAppPrefix is the Apps global namespace.
	KVAppPrefix = ".k"

	// KVUserPrefix is the global namespase used to store OAuth2 user
	// records.
	KVUserPrefix = ".u"

	// KVOAuth2StatePrefix is the global namespase used to store OAuth2
	// ephemeral state data.
	KVOAuth2StatePrefix = ".o"

	// KVSubPrefix is used for keys storing subscriptions.
	KVSubPrefix = "sub."

	// KVInstalledAppPrefix is used to store App records.
	KVInstalledAppPrefix = "app."

	// KVLocalManifestPrefix is used to store locally-listed manifests.
	KVLocalManifestPrefix = "man."

	// KVCallOnceKey and KVClusterMutexKey are used for invoking App Calls once,
	// usually upon a Mattermost instance startup.
	KVCallOnceKey     = "CallOnce"
	KVClusterMutexKey = "Cluster_Mutex"
)
