// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"encoding/json"
	"io"
)

type Subject string

const (
	SubjectUserCreated       = Subject("user_created")
	SubjectUserJoinedChannel = Subject("user_joined_channel")
	SubjectUserLeftChannel   = Subject("user_left_channel")
	SubjectUserJoinedTeam    = Subject("user_joined_team")
	SubjectUserLeftTeam      = Subject("user_left_team")
	SubjectUserUpdated       = Subject("user_updated")
	SubjectChannelCreated    = Subject("channel_created")
	SubjectPostCreated       = Subject("post_created")
)

type Subscription struct {
	AppID   AppID   `json:"app_id"`
	Subject Subject `json:"subject"`

	// Scope
	ChannelID string `json:"channel_id,omitempty"`
	TeamID    string `json:"team_id,omitempty"`

	Call *Call
}

func (sub *Subscription) EqualScope(other *Subscription) bool {
	s1, s2 := *sub, *other
	s1.Call, s2.Call = nil, nil
	return s1 == s2
}

func (sub *Subscription) ToJson() string {
	b, _ := json.Marshal(sub)
	return string(b)
}

type SubscriptionResponse struct {
	Error  string            `json:"error,omitempty"`
	Errors map[string]string `json:"errors,omitempty"`
}

func SubscriptionResponseFromJson(data io.Reader) *SubscriptionResponse {
	var o *SubscriptionResponse
	err := json.NewDecoder(data).Decode(&o)
	if err != nil {
		return nil
	}
	return o
}

func (r *SubscriptionResponse) ToJson() []byte {
	b, _ := json.Marshal(r)
	return b
}
