// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

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

type SubscriptionID string

type Subscription struct {
	SubscriptionID SubscriptionID
	AppID          AppID
	Subject        SubscriptionSubject

	// Scope
	ChannelID string
	TeamID    string
	Regexp    string

	Expand *Expand
}
