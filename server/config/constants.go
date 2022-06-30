// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package config

import (
	"time"
)

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
	WebSocketEventRefreshBindings = "refresh_bindings"
	WebSocketEventPluginEnabled   = "plugin_enabled"
	WebSocketEventPluginDisabled  = "plugin_disabled"
)

const (
	RequestTimeout = time.Second * 30
)

const (
	PropTeamID    = "team_id"
	PropChannelID = "channel_id"
	PropUserAgent = "user_agent_type"
)
