// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"fmt"

	"github.com/hashicorp/go-multierror"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Subject string

const (
	// SubjectUserCreated: system-wide watch for UserHasBeenCreated plugin
	// events.
	//   TeamID: must be empty.
	//   ChannelID must be empty.
	//   Expandable: User.
	//   Requires: model.PermissionViewMembers.
	SubjectUserCreated Subject = "user_created"

	// SubjectUserJoinedChannel, SubjectUserLeftChannel watch the specified
	// channel for users joining and leaving it.
	//   TeamID: must be empty.
	//   ChannelID: specifies the channel to watch.
	//   Expandable: Channel, User, ChannelMember.
	//   Requires: model.PermissionReadChannel permission to ChannelID.
	//
	// To receive and process these notifications as Bot, the bot must first be
	// added as a channel member with an AddChannelMember API call.
	SubjectUserJoinedChannel Subject = "user_joined_channel"
	SubjectUserLeftChannel   Subject = "user_left_channel"

	// SubjectUserJoinedTeam, SubjectUserLeftTeam watch the specified
	// team for users joining and leaving it.
	//   TeamID: specifies the team to watch.
	//   ChannelID: must be empty.
	//   Expandable: Team, User, TeamMember.
	//   Requires: model.PermissionViewTeam
	//
	// To receive and process these notifications as Bot, the bot must first be
	// added as a channel member with an AddTeamMember API call.
	SubjectUserJoinedTeam Subject = "user_joined_team"
	SubjectUserLeftTeam   Subject = "user_left_team"

	// SubjectBotJoinedChannel, SubjectBotLeftChannel watches for the event when
	// the app's own bot is added to, or removed from any channel in the
	// specified team.
	//   TeamID: specifies the team to watch.
	//   ChannelID: must be empty, all channels are watched.
	//   Expandable: Channel, User (will be the app's bot user), ChannelMember.
	//   Requires: none - if the event fires, the app's bot already has the permissions.
	SubjectBotJoinedChannel Subject = "bot_joined_channel"
	SubjectBotLeftChannel   Subject = "bot_left_channel"

	// SubjectBotJoinedTeam, SubjectBotLeftTeam system-wide watch for app's own
	// bot added to, or removed from teams.
	//   TeamID: must be empty.
	//   ChannelID: must be empty.
	//   Expandable: Team, User (will be the app's bot user), TeamMember.
	//   Requires: none - if the event fires, the app's bot already has the permissions.
	SubjectBotJoinedTeam Subject = "bot_joined_team"
	SubjectBotLeftTeam   Subject = "bot_left_team"

	// SubjectChannelCreated watches for new channels in the specified team.
	//   TeamID: specifies the team to watch.
	//   ChannelID: must be empty, all new channels are watched.
	//   Expandable: Channel.
	//   Requires: model.PermissionListTeamChannels.
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
	// Must not contain any extra qualifiers.
	case SubjectUserCreated,
		SubjectBotJoinedTeam,
		SubjectBotLeftTeam /*, SubjectBotMentioned*/ :
		if e.TeamID != "" {
			appendTo = multierror.Append(appendTo, utils.NewInvalidError("teamID must be empty"))
		}
		if e.ChannelID != "" {
			appendTo = multierror.Append(appendTo, utils.NewInvalidError("channelID must be empty"))
		}

	// Require ChannelID, no TeamID
	case SubjectUserJoinedChannel,
		SubjectUserLeftChannel /*, SubjectPostCreated */ :
		if e.TeamID != "" {
			appendTo = multierror.Append(appendTo, utils.NewInvalidError("teamID must be empty"))
		}
		if e.ChannelID == "" {
			appendTo = multierror.Append(appendTo, utils.NewInvalidError("channelID must not be empty"))
		}

	// Require TeamID, no ChannelID
	case SubjectUserJoinedTeam,
		SubjectUserLeftTeam,
		SubjectBotJoinedChannel,
		SubjectBotLeftChannel,
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

func (e Event) Loggable() []interface{} {
	props := []interface{}{"subject", string(e.Subject)}
	if e.ChannelID != "" {
		props = append(props, "channel_id", e.ChannelID)
	}
	if e.TeamID != "" {
		props = append(props, "team_id", e.TeamID)
	}
	return props
}

func (e Event) String() string {
	s := fmt.Sprintf("subject: %s", e.Subject)
	if e.ChannelID != "" {
		s += fmt.Sprintf(", channel_id: %s", e.ChannelID)
	}
	if e.TeamID != "" {
		s += fmt.Sprintf(", team_id: %s", e.TeamID)
	}
	return s
}
