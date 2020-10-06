// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import "github.com/mattermost/mattermost-plugin-apps/server/constants"

type SubscriptionID string

type Subscription struct {
	SubscriptionID SubscriptionID
	AppID          AppID
	Subject        constants.SubscriptionSubject

	// Scope
	ChannelID string
	TeamID    string
	Regexp    string

	Expand *Expand
}
