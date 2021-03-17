// +build !e2e

package apps

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshalCallRequest(t *testing.T) {
	const payload = `
	{
		"context": {
			"acting_user_id": "q45j6a851fgr98iqr3mdxx3cye",
			"team_id": "9pu8hstcpigm5x4dboe6hz9ddw",
			"mattermost_site_url": "https://levb.ngrok.io"
		},
		"values": {
			"oauth2_client_secret": "cywc3e8nebyujrpuip98t69a3h",
			"selected_option": {
				"label": "The Label",
				"value": "The Value"
			}
		}
	}
	`

	data, err := UnmarshalCallRequestFromData([]byte(payload))

	require.NoError(t, err)
	require.Equal(t, "q45j6a851fgr98iqr3mdxx3cye", data.Context.ActingUserID)
	require.Equal(t, "9pu8hstcpigm5x4dboe6hz9ddw", data.Context.TeamID)
	require.Equal(t, "https://levb.ngrok.io", data.Context.MattermostSiteURL)
	require.Equal(t, "cywc3e8nebyujrpuip98t69a3h", data.Values[PropOAuth2ClientSecret])
	require.Equal(t, "The Value", data.GetValue("selected_option", ""))
	require.Equal(t, "The Default Value", data.GetValue("nonexistent", "The Default Value"))
}

func TestMarshalCallResponse(t *testing.T) {
	resStr := `{
		"type": "form",
		"form": {
			"fields": [
				{
					"name": "field1",
					"value": "value1"
				}
			]
		}
	}`
	res := &CallResponse{}

	err := json.Unmarshal([]byte(resStr), res)
	require.NoError(t, err)

	data, err := json.Marshal(res.Form.Fields[0])
	require.NoError(t, err)

	m := map[string]string{}
	err = json.Unmarshal(data, &m)

	require.NoError(t, err)
	require.Equal(t, "field1", m["name"])
	require.Equal(t, "value1", m["value"])
}
