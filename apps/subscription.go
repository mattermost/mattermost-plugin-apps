// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"encoding/json"
	"io"
)

type Subject string

const (
	// SubjectUserCreated subscribes to UserHasBeenCreated plugin events. By
	// default, fully expanded User object is included in the notifications.
	// There is no other data to expand.
	SubjectUserCreated Subject = "user_created"

	// SubjectUserJoinedChannel and SubjectUserLeftChannel subscribes to
	// respective plugin events, for the specified channel. By default
	// notifications include ActingUserID, UserID, and ChannelID, but only
	// ActingUser is fully expanded. Expand can be used to expand other
	// entities.
	SubjectUserJoinedChannel Subject = "user_joined_channel"
	SubjectUserLeftChannel   Subject = "user_left_channel"
	SubjectBotJoinedChannel  Subject = "bot_joined_channel"
	SubjectBotLeftChannel    Subject = "bot_left_channel"

	// SubjectUserJoinedTeam and SubjectUserLeftTeam subscribes to respective
	// plugin events, for the specified team. By default notifications include
	// ActingUserID, UserID, and TeamID, but only ActingUser is fully expanded.
	// Expand can be used to expand other entities.
	SubjectUserJoinedTeam Subject = "user_joined_team"
	SubjectUserLeftTeam   Subject = "user_left_team"
	SubjectBotJoinedTeam  Subject = "bot_joined_team"
	SubjectBotLeftTeam    Subject = "bot_left_team"

	// SubjectChannelCreated subscribes to ChannelHasBeenCreated plugin events,
	// for the specified team. By default notifications include UserID (creator),
	// ChannelID, and TeamID, but only Channel is fully expanded. Expand can be
	// used to expand other entities.
	SubjectChannelCreated Subject = "channel_created"

	// SubjectPostCreated subscribes to MessageHasBeenPosted plugin events, for
	// the specified channel. By default notifications include UserID (author), PostID,
	// RootPostID, ChannelID, but only Post is fully expanded. Expand can be
	// used to expand other entities.
	SubjectPostCreated Subject = "post_created"

	// SubjectBotMentioned subscribes to MessageHasBeenPosted plugin events, specifically
	// when the App's bot is mentioned in the post.
	SubjectBotMentioned Subject = "bot_mentioned"
)

// Subscription is submitted by an app to the Subscribe API. It determines what
// events the app would like to be notified on, and how these notifications
// should be invoked.
type Subscription struct {
	// AppID is used internally by Mattermost. It does not need to be set by app
	// developers.
	AppID AppID `json:"app_id,omitempty"`

	// Subscription subject. See type Subject godoc (linked) for details.
	Subject Subject `json:"subject"`

	// ChannelID and TeamID are the subscription scope, as applicable to the subject.
	ChannelID string `json:"channel_id,omitempty"`
	TeamID    string `json:"team_id,omitempty"`

	// Call is the (one-way) call to make upon the event.
	Call *Call
}

func (sub *Subscription) EqualScope(other *Subscription) bool {
	s1, s2 := *sub, *other
	s1.Call, s2.Call = nil, nil
	return s1 == s2
}

func (sub *Subscription) ToJSON() string {
	b, _ := json.Marshal(sub)
	return string(b)
}

type SubscriptionResponse struct {
	Error  string            `json:"error,omitempty"`
	Errors map[string]string `json:"errors,omitempty"`
}

func SubscriptionResponseFromJSON(data io.Reader) *SubscriptionResponse {
	var o *SubscriptionResponse
	err := json.NewDecoder(data).Decode(&o)
	if err != nil {
		return nil
	}
	return o
}

func (r *SubscriptionResponse) ToJSON() []byte {
	b, _ := json.Marshal(r)
	return b
}
