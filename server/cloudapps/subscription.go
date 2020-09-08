// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package cloudapps

type SubscriptionSubject string

const (
	SubjectUserJoinedChannel = SubscriptionSubject("user_joined_channel")
)

type SubscriptionID string

type Subscription struct {
	SubscriptionID SubscriptionID
	AppID          AppID
	Subject        SubscriptionSubject

	// Scope
	ChannelID string
	Expand    *Expand
}
