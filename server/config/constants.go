// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package config

import "github.com/mattermost/mattermost-plugin-apps/apps"

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

	// OAuth2 App's HTTP endpoints in the {PluginURL}/apps/{AppID} space.
	PathMattermostOAuth2Connect  = "/oauth2/mattermost/connect"
	PathMattermostOAuth2Complete = "/oauth2/mattermost/complete"
	PathRemoteOAuth2Connect      = "/oauth2/remote/connect"
	PathRemoteOAuth2Complete     = "/oauth2/remote/complete"

	// Static assets are served from {PluginURL}/static/...
	PathStatic = "/" + apps.StaticFolder

	// Marketplace sub-paths.
	PathMarketplace = "/marketplace"

	PathInvalidateCache = "/invalidatecache"

	WebSocketEventRefreshBindings = "refresh_bindings"
)

const (
	PropTeamID    = "team_id"
	PropChannelID = "channel_id"
	PropPostID    = "post_id"
	PropUserAgent = "user_agent_type"
)

// KV namespace. The use of '.' in the prefixes is to avoid conflicts with
// base64 URL encoding that already uses '-' and '_'.
const (
	// KVAppPrefix is the Apps namespace. Short, maximize the app keyspace
	KVAppPrefix = "kv."

	// KVOAuth2Prefix is used to store OAuth2-related information (state,
	// tokens)
	KVOAuth2Prefix      = "o."
	KVOAuth2StatePrefix = "s."

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
