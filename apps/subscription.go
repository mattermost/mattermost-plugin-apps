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

	// SubjectUserJoinedChannel, SubjectUserLeftChannel behave differently
	// depending on the provided parameters. If no channel is specified, the
	// subscription is system-wide, and the event is fired for every channel the
	// subscriber joins or leaves. If a channel is specified (and the user has
	// access to the channel), the event is fired for all users joining or
	// leaving the channel.
	//   TeamID: must be empty.
	//   ChannelID: (optional) specifies the channel to watch.
	//   Expandable: Channel, Team, User, ChannelMember.
	//   Requires: model.PermissionReadChannel permission to ChannelID.
	//
	// TODO: add channel member as part of the subscribe call for the app's bot user?
	// To receive and process these notifications as Bot, the bot must first be
	// added as a channel member with an AddChannelMember API call.
	SubjectUserJoinedChannel Subject = "user_joined_channel"
	SubjectUserLeftChannel   Subject = "user_left_channel"

	// SubjectUserJoinedTeam, SubjectUserLeftTeam behave differently
	// depending on the provided parameters. If no team is specified, the
	// subscription is system-wide, and the event is fired for every team the
	// subscriber joins or leaves. If a team is specified (and the user has
	// access to the team), the event is fired for all users joining or
	// leaving the team.
	//   TeamID: (optional) specifies the team to watch.
	//   ChannelID: must be empty.
	//   Expandable: Team, User, TeamMember.
	//   Requires: model.PermissionViewTeam permission to TeamID.
	//
	// TODO: add team member as part of the subscribe call for the app's bot user?
	// To receive and process these notifications as Bot, the bot must first be
	// added as a team member with an AddTeamMember API call.
	SubjectUserJoinedTeam Subject = "user_joined_team"
	SubjectUserLeftTeam   Subject = "user_left_team"

	// SubjectBotJoinedChannelDeprecated, SubjectBotLeftChannel,
	// SubjectBotJoinedTeam, SubjectBotLeftTeam are deprecated. Use "User"
	// instead. They are kept for backwards compatibility, and only work in the
	// "system-wide" mode for the app's bot user.
	SubjectBotJoinedChannelDeprecated Subject = "bot_joined_channel"
	SubjectBotLeftChannelDeprecated   Subject = "bot_left_channel"
	SubjectBotJoinedTeamDeprecated    Subject = "bot_joined_team"
	SubjectBotLeftTeamDeprecated      Subject = "bot_left_team"

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

	// SubjectSelfMentioned subscribes to MessageHasBeenPosted plugin events, specifically
	// when the subscriber is mentioned in the post.
	// SubjectSelfMentioned Subject = "self_mentioned"
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
		result = multierror.Append(result, utils.NewInvalidError("call must not be empty"))
	}
	return sub.Event.validate(result)
}

func (e Event) Validate() error {
	return e.validate(nil)
}

func (e Event) validate(appendTo error) error {
	if e.Subject == "" {
		appendTo = multierror.Append(appendTo, utils.NewInvalidError("subject must not be empty"))
	}

	switch e.Subject {
	// Globally scoped, must not contain any extra qualifiers.
	case SubjectUserCreated,
		SubjectBotJoinedTeamDeprecated,
		SubjectBotLeftTeamDeprecated,
		SubjectBotJoinedChannelDeprecated,
		SubjectBotLeftChannelDeprecated /*, SubjectSelfMentioned*/ :
		if e.TeamID != "" {
			appendTo = multierror.Append(appendTo, utils.NewInvalidError("%s is scoped globally; team_id and channel_id must both be empty", e.Subject))
		}
		if e.ChannelID != "" {
			appendTo = multierror.Append(appendTo, utils.NewInvalidError("%s is scoped globally; team_id and channel_id must both be empty", e.Subject))
		}

	// Team scoped, require TeamID, no ChannelID.
	case SubjectChannelCreated:
		if e.TeamID == "" {
			appendTo = multierror.Append(appendTo, utils.NewInvalidError("%s is scoped to a team; team_id must not be empty", e.Subject))
		}
		if e.ChannelID != "" {
			appendTo = multierror.Append(appendTo, utils.NewInvalidError("%s is scoped to a team; channel_id must be empty", e.Subject))
		}

	// Special case SubjectUserJoinedTeam, SubjectUserLeftTeam: optional TeamID,
	// no ChannelID.
	case SubjectUserJoinedTeam, SubjectUserLeftTeam:
		if e.ChannelID != "" {
			appendTo = multierror.Append(appendTo, utils.NewInvalidError("%s is scoped globally, or to a team; channel_id must be empty", e.Subject))
		}

	// Special case SubjectUserJoinedChannel, SubjectUserLeftChannel: optional ChannelID,
	// no TeamID.
	case SubjectUserJoinedChannel, SubjectUserLeftChannel:
		if e.TeamID != "" {
			appendTo = multierror.Append(appendTo, utils.NewInvalidError("%s is scoped globally, or to a channel; team_id must be empty", e.Subject))
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
