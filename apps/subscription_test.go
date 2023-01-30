// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func TestValidateSubscriptionEvent(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		apps.Event
		expectedError string
	}{
		// Good.
		// {Event: apps.Event{Subject: apps.SubjectUserCreated}},
		{Event: apps.Event{Subject: apps.SubjectUserJoinedChannel, ChannelID: "channelID"}},
		{Event: apps.Event{Subject: apps.SubjectUserLeftChannel, ChannelID: "channelID"}},
		{Event: apps.Event{Subject: apps.SubjectUserJoinedTeam, TeamID: "teamID"}},
		{Event: apps.Event{Subject: apps.SubjectUserLeftTeam, TeamID: "teamID"}},
		{Event: apps.Event{Subject: apps.SubjectUserJoinedChannel}},
		{Event: apps.Event{Subject: apps.SubjectUserLeftChannel}},
		{Event: apps.Event{Subject: apps.SubjectUserJoinedTeam}},
		{Event: apps.Event{Subject: apps.SubjectUserLeftTeam}},
		{Event: apps.Event{Subject: apps.SubjectBotJoinedChannel_Deprecated}},
		{Event: apps.Event{Subject: apps.SubjectBotLeftChannel_Deprecated}},
		{Event: apps.Event{Subject: apps.SubjectBotJoinedTeam_Deprecated}},
		{Event: apps.Event{Subject: apps.SubjectBotLeftTeam_Deprecated}},
		{Event: apps.Event{Subject: apps.SubjectChannelCreated, TeamID: "teamID"}},

		// Bad.
		{
			Event:         apps.Event{Subject: apps.SubjectUserCreated, TeamID: "teamID"},
			expectedError: "user_created is scoped globally; team_id and channel_id must both be empty: invalid input",
		},
		{
			Event:         apps.Event{Subject: apps.SubjectUserJoinedChannel, TeamID: "teamID", ChannelID: "channelID"},
			expectedError: "user_joined_channel is scoped globally, or to a channel; team_id must be empty: invalid input",
		},
		{
			Event:         apps.Event{Subject: apps.SubjectUserJoinedChannel, TeamID: "teamID"},
			expectedError: "user_joined_channel is scoped globally, or to a channel; team_id must be empty: invalid input",
		},
		{
			Event:         apps.Event{Subject: apps.SubjectUserLeftChannel, TeamID: "teamID", ChannelID: "channelID"},
			expectedError: "user_left_channel is scoped globally, or to a channel; team_id must be empty: invalid input",
		},
		{
			Event:         apps.Event{Subject: apps.SubjectUserLeftChannel, TeamID: "teamID"},
			expectedError: "user_left_channel is scoped globally, or to a channel; team_id must be empty: invalid input",
		},
		{
			Event:         apps.Event{Subject: apps.SubjectUserJoinedTeam, TeamID: "teamID", ChannelID: "channelID"},
			expectedError: "user_joined_team is scoped globally, or to a team; channel_id must be empty: invalid input",
		},
		{
			Event:         apps.Event{Subject: apps.SubjectUserJoinedTeam, ChannelID: "channelID"},
			expectedError: "user_joined_team is scoped globally, or to a team; channel_id must be empty: invalid input",
		},
		{
			Event:         apps.Event{Subject: apps.SubjectUserLeftTeam, TeamID: "teamID", ChannelID: "channelID"},
			expectedError: "user_left_team is scoped globally, or to a team; channel_id must be empty: invalid input",
		},
		{
			Event:         apps.Event{Subject: apps.SubjectUserLeftTeam, ChannelID: "channelID"},
			expectedError: "user_left_team is scoped globally, or to a team; channel_id must be empty: invalid input",
		},
		{
			Event:         apps.Event{Subject: apps.SubjectBotJoinedChannel_Deprecated, TeamID: "teamID"},
			expectedError: "bot_joined_channel is scoped globally; team_id and channel_id must both be empty: invalid input",
		},
		{
			Event:         apps.Event{Subject: apps.SubjectBotLeftChannel_Deprecated, TeamID: "teamID"},
			expectedError: "bot_left_channel is scoped globally; team_id and channel_id must both be empty: invalid input",
		},
		{
			Event:         apps.Event{Subject: apps.SubjectBotJoinedTeam_Deprecated, TeamID: "teamID"},
			expectedError: "bot_joined_team is scoped globally; team_id and channel_id must both be empty: invalid input",
		},
		{
			Event:         apps.Event{Subject: apps.SubjectBotLeftTeam_Deprecated, TeamID: "teamID"},
			expectedError: "bot_left_team is scoped globally; team_id and channel_id must both be empty: invalid input",
		},
		{
			Event:         apps.Event{Subject: apps.SubjectChannelCreated},
			expectedError: "channel_created is scoped to a team; team_id must not be empty: invalid input",
		},
		{
			Event:         apps.Event{Subject: apps.SubjectChannelCreated, TeamID: "teamID", ChannelID: "channelID"},
			expectedError: "channel_created is scoped to a team; channel_id must be empty: invalid input",
		},
	} {
		t.Run(tc.Event.String(), func(t *testing.T) {
			err := tc.Event.Validate()
			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, "1 error occurred:\n\t* "+tc.expectedError+"\n\n", err.Error())
			}
		})
	}
}
