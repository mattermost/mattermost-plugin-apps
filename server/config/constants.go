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

	// Top-level path for the REST APIs exposed by the plugin itself.
	APIPath = "/api/v1"

	// Top-level path for the Apps namespaces, followed by /AppID/subpath.
	AppsPath = "/apps"

	// OAuth2 sub-paths.
	PathOAuth2             = "/oauth2"
	PathMattermostRedirect = "/mattermost/redirect"
	PathMattermostComplete = "/mattermost/complete"
	PathRemoteRedirect     = "/remote/redirect"
	PathRemoteComplete     = "/remote/complete"

	// Marketplace sub-paths.
	PathMarketplace = "/marketplace"

	// Other sub-paths.
	CallPath        = "/call"
	KVPath          = "/kv"
	SubscribePath   = "/subscribe"
	UnsubscribePath = "/unsubscribe"
	StaticAssetPath = "/" + apps.StaticAssetsFolder

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
	KVAppPrefix = "a_"

	// PrefixOAuth2 is used to store OAuth2-related information (state, tokens)
	KVOAuth2Prefix = "oauth2_"

	// KVCallOnceKey and KVClusterMutexKey are used for invoking App Calls once,
	// usually upon a Mattermost instance startup.
	KVCallOnceKey     = "CallOnce"
	KVClusterMutexKey = "Cluster_Mutex"

	// PrefixSub is used for keys storing subscriptions.
	KVSubPrefix = "sub_"

	// PrefixInstalledApp is used to store App records.
	KVInstalledAppPrefix = "app_"

	// PrefixLocalManifest is used to store locally-listed manifests.
	KVLocalManifestPrefix = "man_"
)
