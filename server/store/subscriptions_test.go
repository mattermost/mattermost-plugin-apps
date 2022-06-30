package store

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func TestSubsKey(t *testing.T) {
	for subject, expected := range map[apps.Subject]string{
		apps.SubjectUserCreated:       "sub.user_created",
		apps.SubjectUserJoinedChannel: "sub.user_joined_channel.channel-id",
		apps.SubjectUserLeftChannel:   "sub.user_left_channel.channel-id",
		apps.SubjectUserJoinedTeam:    "sub.user_joined_team.team-id",
		apps.SubjectUserLeftTeam:      "sub.user_left_team.team-id",
		apps.SubjectChannelCreated:    "sub.channel_created.team-id",
		// apps.SubjectPostCreated: "sub.post_created.channel-id",
	} {
		t.Run(string(subject), func(t *testing.T) {
			r, err := subsKey(apps.Event{
				Subject:   subject,
				TeamID:    "team-id",
				ChannelID: "channel-id",
			})
			require.NoError(t, err)
			require.Equal(t, expected, r)
		})
	}
}
