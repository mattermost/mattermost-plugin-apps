package goapp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func TestCallRequestJSON(t *testing.T) {
	creq := apps.CallRequest{
		Call: *apps.NewCall("/som-path").
			WithExpand(apps.Expand{
				ActingUser: apps.ExpandAll,
			}),
		Context: apps.Context{
			ExpandedContext: apps.ExpandedContext{
				User: &model.User{
					Id:       "abcdefghijklmnopqrstuvwxyz",
					Username: "test",
				},
			},
		},
		Values: map[string]interface{}{
			"testkey": "testvalue",
		},
	}
	json1, err := json.Marshal(creq)
	require.NoError(t, err)

	goappReq := CallRequest{
		CallRequest: creq,
	}
	json2, err := json.Marshal(goappReq)
	require.NoError(t, err)
	require.Equal(t, json1, json2)
}
