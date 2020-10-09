// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package constants

const (
	Repository     = "mattermost-plugin-apps"
	CommandTrigger = "apps"
)

const (
	BotUsername    = "appsbot"
	BotDisplayName = "Mattermost Apps"
	BotDescription = "Mattermost Apps Registry and API proxy."

	APIPath               = "/api/v1"
	InteractiveDialogPath = "/dialog"
	HelloAppPath          = "/hello"
)

type SubscriptionSubject string

const (
	SubjectUserCreated       = SubscriptionSubject("user_created")
	SubjectUserJoinedChannel = SubscriptionSubject("user_joined_channel")
	SubjectUserLeftChannel   = SubscriptionSubject("user_left_channel")
	SubjectUserJoinedTeam    = SubscriptionSubject("user_joined_team")
	SubjectUserLeftTeam      = SubscriptionSubject("user_left_team")
	SubjectUserUpdated       = SubscriptionSubject("user_updated")
	SubjectChannelCreated    = SubscriptionSubject("channel_created")
	SubjectPostCreated       = SubscriptionSubject("post_created")
)
