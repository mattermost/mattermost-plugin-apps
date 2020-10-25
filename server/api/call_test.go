package api

import (
	"testing"

	"github.com/mattermost/mattermost-plugin-apps/server/constants"

	"github.com/stretchr/testify/require"
)

func TestUnmarshalCallRequest(t *testing.T) {
	const payload = `
	{
		"context": {
			"acting_user_id": "q45j6a851fgr98iqr3mdxx3cye",
			"team_id": "9pu8hstcpigm5x4dboe6hz9ddw",
			"config": {
				"site_url": "https://levb.ngrok.io"
			}
		},
		"values": {
			"bot_access_token": "b3snp6tk6pbh9fxjpbhqn5hzgh",
			"oauth2_client_secret": "cywc3e8nebyujrpuip98t69a3h"
		}
	}
	`

	data, err := UnmarshalCallFromData([]byte(payload))

	require.NoError(t, err)
	require.Equal(t, "q45j6a851fgr98iqr3mdxx3cye", data.Context.ActingUserID)
	require.Equal(t, "9pu8hstcpigm5x4dboe6hz9ddw", data.Context.TeamID)
	require.Equal(t, "https://levb.ngrok.io", data.Context.Config.SiteURL)
	require.Equal(t, "b3snp6tk6pbh9fxjpbhqn5hzgh", data.Values[constants.BotAccessToken])
	require.Equal(t, "cywc3e8nebyujrpuip98t69a3h", data.Values[constants.OAuth2ClientSecret])
}
