package store

import (
	"testing"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/stretchr/testify/require"
)

func Test_subsKey(t *testing.T) {
	type testcase struct {
		e        apps.Event
		expected string
	}

	tcs := func(subject apps.Subject, expectedNone string, expectedChannelID string, expectedTeamID string, expectedBoth string) []testcase {
		return []testcase{
			{
				e:        apps.Event{Subject: subject},
				expected: expectedNone,
			},
			{
				e:        apps.Event{Subject: subject, ChannelID: "channelID"},
				expected: expectedChannelID,
			},
			{
				e:        apps.Event{Subject: subject, TeamID: "teamID"},
				expected: expectedTeamID,
			},
			{
				e:        apps.Event{Subject: subject, ChannelID: "channelID", TeamID: "teamID"},
				expected: expectedBoth,
			},
		}
	}

	tests := []testcase{}
	tests = append(tests, tcs(apps.SubjectChannelCreated,
		"can't make a key for a subscription, expected a team ID for subject channel_created",
		"can't make a key for a subscription, expected channel ID empty for subject channel_created",
		"sub.channel_created.teamID",
		"can't make a key for a subscription, expected channel ID empty for subject channel_created")...)

	tests = append(tests, tcs(apps.SubjectUserCreated,
		"sub.user_created",
		"can't make a key for a subscription, expected team and channel IDs empty for subject user_created",
		"can't make a key for a subscription, expected team and channel IDs empty for subject user_created",
		"can't make a key for a subscription, expected team and channel IDs empty for subject user_created")...)

	tests = append(tests, tcs(apps.SubjectUserJoinedChannel,
		"sub.user_joined_channel",
		"sub.user_joined_channel.channelID",
		"can't make a key for a subscription, expected team ID empty for subject user_joined_channel",
		"can't make a key for a subscription, expected team ID empty for subject user_joined_channel")...)

	tests = append(tests, tcs(apps.SubjectUserLeftChannel,
		"sub.user_left_channel",
		"sub.user_left_channel.channelID",
		"can't make a key for a subscription, expected team ID empty for subject user_left_channel",
		"can't make a key for a subscription, expected team ID empty for subject user_left_channel")...)

	tests = append(tests, tcs(apps.SubjectUserJoinedTeam,
		"sub.user_joined_team",
		"can't make a key for a subscription, expected channel ID empty for subject user_joined_team",
		"sub.user_joined_team.teamID",
		"can't make a key for a subscription, expected channel ID empty for subject user_joined_team")...)

	tests = append(tests, tcs(apps.SubjectUserLeftTeam,
		"sub.user_left_team",
		"can't make a key for a subscription, expected channel ID empty for subject user_left_team",
		"sub.user_left_team.teamID",
		"can't make a key for a subscription, expected channel ID empty for subject user_left_team")...)

	tests = append(tests, tcs(apps.SubjectBotJoinedChannel_Deprecated,
		"sub.bot_joined_channel",
		"can't make a key for a subscription, expected team and channel IDs empty for subject bot_joined_channel",
		"can't make a key for a subscription, expected team and channel IDs empty for subject bot_joined_channel",
		"can't make a key for a subscription, expected team and channel IDs empty for subject bot_joined_channel")...)

	tests = append(tests, tcs(apps.SubjectBotLeftChannel_Deprecated,
		"sub.bot_left_channel",
		"can't make a key for a subscription, expected team and channel IDs empty for subject bot_left_channel",
		"can't make a key for a subscription, expected team and channel IDs empty for subject bot_left_channel",
		"can't make a key for a subscription, expected team and channel IDs empty for subject bot_left_channel")...)

	tests = append(tests, tcs(apps.SubjectBotJoinedTeam_Deprecated,
		"sub.bot_joined_team",
		"can't make a key for a subscription, expected team and channel IDs empty for subject bot_joined_team",
		"can't make a key for a subscription, expected team and channel IDs empty for subject bot_joined_team",
		"can't make a key for a subscription, expected team and channel IDs empty for subject bot_joined_team")...)

	tests = append(tests, tcs(apps.SubjectBotLeftTeam_Deprecated,
		"sub.bot_left_team",
		"can't make a key for a subscription, expected team and channel IDs empty for subject bot_left_team",
		"can't make a key for a subscription, expected team and channel IDs empty for subject bot_left_team",
		"can't make a key for a subscription, expected team and channel IDs empty for subject bot_left_team")...)

	for _, tc := range tests {
		t.Run(tc.e.String(), func(t *testing.T) {
			got, err := subsKey(tc.e)
			if err != nil {
				got = err.Error()
			}
			require.Equal(t, tc.expected, got)
		})
	}
}
