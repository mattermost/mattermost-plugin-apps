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
	PathMattermostOAuth2Redirect = "/oauth2/mattermost/redirect"
	PathMattermostOAuth2Complete = "/oauth2/mattermost/complete"
	PathRemoteOAuth2Redirect     = "/oauth2/remote/redirect"
	PathRemoteOAuth2Complete     = "/oauth2/remote/complete"

	// Static assets are served from {PluginURL}/static/...
	PathStatic = "/" + apps.StaticFolder

	// Marketplace sub-paths.
	PathMarketplace = "/marketplace"

	WebSocketEventRefreshBindings = "refresh_bindings"
)

const (
	PropTeamID    = "team_id"
	PropChannelID = "channel_id"
	PropPostID    = "post_id"
	PropUserAgent = "user_agent_type"
)

// KV namespace
const (
	// PrefixApp is the Apps namespace. Short, maximize the app keyspace
	KVAppPrefix = "_"

	// PrefixOAuth2 is used to store OAuth2-related information (state, tokens)
	KVOAuth2Prefix = "oauth2_"

	// PrefixSub is used for keys storing subscriptions.
	KVSubPrefix = "sub_"

	// PrefixInstalledApp is used to store App records.
	KVInstalledAppPrefix = "app_"

	// PrefixLocalManifest is used to store locally-listed manifests.
	KVLocalManifestPrefix = "man_"

	// KVCallOnceKey and KVClusterMutexKey are used for invoking App Calls once,
	// usually upon a Mattermost instance startup.
	KVCallOnceKey     = "CallOnce"
	KVClusterMutexKey = "Cluster_Mutex"
)
