// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

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

	Expand *Expand `json:"expand,omitempty"`
}

type Notification struct {
	Subject Subject
	Context *Context
}

func (sub *Subscription) EqualScope(other *Subscription) bool {
	s1, s2 := *sub, *other
	s1.Expand, s2.Expand = nil, nil
	return s1 == s2
}
