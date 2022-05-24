// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/hashicorp/go-multierror"

	"github.com/mattermost/mattermost-plugin-apps/utils"
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

	// SubjectBotJoinedChannel and SubjectBotLeftChannel subscribes to
	// SubjectUserJoinedChannel and SubjectUserLeftChannel plugin events,
	// specifically when the App's bot is added or removed from the channel
	SubjectBotJoinedChannel Subject = "bot_joined_channel"
	SubjectBotLeftChannel   Subject = "bot_left_channel"

	// SubjectUserJoinedTeam and SubjectUserLeftTeam subscribes to respective
	// plugin events, for the specified team. By default notifications include
	// ActingUserID, UserID, and TeamID, but only ActingUser is fully expanded.
	// Expand can be used to expand other entities.
	SubjectUserJoinedTeam Subject = "user_joined_team"
	SubjectUserLeftTeam   Subject = "user_left_team"

	// SubjectBotJoinedTeam and SubjectBotLeftTeam subscribes to
	// SubjectUserJoinedTeam and SubjectUserLeftTeam plugin events,
	// specifically when the App's bot is added or removed from the team
	SubjectBotJoinedTeam Subject = "bot_joined_team"
	SubjectBotLeftTeam   Subject = "bot_left_team"

	// SubjectChannelCreated subscribes to ChannelHasBeenCreated plugin events,
	// for the specified team. By default notifications include UserID (creator),
	// ChannelID, and TeamID, but only Channel is fully expanded. Expand can be
	// used to expand other entities.
	SubjectChannelCreated Subject = "channel_created"

	// TODO: re-enable post_created and bot_mentioned once perf issues are
	// resolved, see https://mattermost.atlassian.net/browse/MM-44388

	// SubjectPostCreated subscribes to MessageHasBeenPosted plugin events, for
	// the specified channel. By default notifications include UserID (author), PostID,
	// RootPostID, ChannelID, but only Post is fully expanded. Expand can be
	// used to expand other entities.
	// SubjectPostCreated Subject = "post_created"

	// SubjectBotMentioned subscribes to MessageHasBeenPosted plugin events, specifically
	// when the App's bot is mentioned in the post.
	// SubjectBotMentioned Subject = "bot_mentioned"
)

// Subscription is submitted by an app to the Subscribe API. It determines what
// events the app would like to be notified on, and how these notifications
// should be invoked.
type Subscription struct {
	Event

	// Call is the (one-way) call to make upon the event.
	Call Call `json:"call"`
}

type Event struct {
	// Subscription subject. See type Subject godoc (linked) for details.
	Subject Subject `json:"subject"`

	// ChannelID and TeamID are the subscription scope, as applicable to the subject.
	ChannelID string `json:"channel_id,omitempty"`
	TeamID    string `json:"team_id,omitempty"`
}

func (sub Subscription) Validate() error {
	var result error
	emptyCall := Call{}
	if sub.Call == emptyCall {
		result = multierror.Append(result, utils.NewInvalidError("call most not be empty"))
	}
	return sub.Event.validate(result)
}

func (e Event) Validate() error {
	var result error
	return e.validate(result)
}

func (e Event) validate(appendTo error) error {
	if e.Subject == "" {
		appendTo = multierror.Append(appendTo, utils.NewInvalidError("subject most not be empty"))
	}

	switch e.Subject {
	case SubjectUserCreated,
		SubjectBotJoinedChannel,
		SubjectBotLeftChannel,
		SubjectBotJoinedTeam,
		SubjectBotLeftTeam /*, SubjectBotMentioned*/ :
		if e.TeamID != "" {
			appendTo = multierror.Append(appendTo, utils.NewInvalidError("teamID must be empty"))
		}
		if e.ChannelID != "" {
			appendTo = multierror.Append(appendTo, utils.NewInvalidError("channelID must be empty"))
		}

	case SubjectUserJoinedChannel,
		SubjectUserLeftChannel /*, SubjectPostCreated */ :
		if e.TeamID != "" {
			appendTo = multierror.Append(appendTo, utils.NewInvalidError("teamID must be empty"))
		}

		if e.ChannelID == "" {
			appendTo = multierror.Append(appendTo, utils.NewInvalidError("channelID must not be empty"))
		}

	case SubjectUserJoinedTeam,
		SubjectUserLeftTeam,
		SubjectChannelCreated:
		if e.TeamID == "" {
			appendTo = multierror.Append(appendTo, utils.NewInvalidError("teamID must not be empty"))
		}

		if e.ChannelID != "" {
			appendTo = multierror.Append(appendTo, utils.NewInvalidError("channelID must be empty"))
		}
	default:
		appendTo = multierror.Append(appendTo, utils.NewInvalidError("Unknown subject %s", e.Subject))
	}

	return appendTo
}

// EqualScope compares 2 subscriptions to have the same scope, i.e. Subject, and
// Channel/Team IDs.
func (sub Subscription) EqualScope(s2 Subscription) bool {
	sub.Call, s2.Call = Call{}, Call{}
	return sub == s2
}

func (sub Subscription) Loggable() []interface{} {
	props := []interface{}{"subject", sub.Subject}
	if len(sub.ChannelID) > 0 {
		props = append(props, "channel_id", sub.ChannelID)
	}
	if len(sub.TeamID) > 0 {
		props = append(props, "team_id", sub.TeamID)
	}
	return props
}
