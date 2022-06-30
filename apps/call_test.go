// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func TestUnmarshalCall(t *testing.T) {
	const full = `
	{
		"path": "/test",
		"state": {
			"key": "value"
		},
		"expand": {
			"acting_user": "all"
		}
	}
	`

	c := apps.NewCall("")
	err := json.Unmarshal([]byte(full), c)
	require.NoError(t, err)
	require.Equal(t, &apps.Call{
		Path: "/test",
		State: map[string]interface{}{
			"key": "value",
		},
		Expand: &apps.Expand{
			ActingUser: apps.ExpandAll,
		},
	}, c)

	const short = `"/test"`
	c = apps.NewCall("")
	err = json.Unmarshal([]byte(short), c)
	require.NoError(t, err)
	require.Equal(t, apps.NewCall("/test"), c)
}

func TestUnmarshalCallRequest(t *testing.T) {
	const payload = `
	{
		"context": {
			"team_id": "9pu8hstcpigm5x4dboe6hz9ddw",
			"mattermost_site_url": "https://some.test"
		},
		"values": {
			"secret": "cywc3e8nebyujrpuip98t69a3h",
			"selected_option": {
				"label": "The Label",
				"value": "The Value"
			}
		}
	}
	`

	data, err := apps.CallRequestFromJSON([]byte(payload))

	require.NoError(t, err)
	require.Equal(t, "9pu8hstcpigm5x4dboe6hz9ddw", data.Context.TeamID)
	require.Equal(t, "https://some.test", data.Context.MattermostSiteURL)
	require.Equal(t, "cywc3e8nebyujrpuip98t69a3h", data.Values["secret"])
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
	res := &apps.CallResponse{}

	err := json.Unmarshal([]byte(resStr), res)
	require.NoError(t, err)

	data, err := json.Marshal(res.Form.Fields[0])
	require.NoError(t, err)

	m := map[string]interface{}{}
	err = json.Unmarshal(data, &m)

	require.NoError(t, err)
	require.Equal(t, "field1", m["name"])
	require.Equal(t, "value1", m["value"])
}
