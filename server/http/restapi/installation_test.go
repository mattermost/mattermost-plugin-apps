// +build app

package restapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/http/dialog"
)

func TestAppInstallCommand(t *testing.T) {
	th := Setup(t)
	SetupPP(th, t)
	defer th.TearDown()

	appID := "test"
	channelID, err := th.Client.GetChannelID()

	require.NoError(t, err)

	command := fmt.Sprintf("/apps install --app-id %s", appID)
	response, err := th.Client.ExecuteCommand(channelID, command)
	require.NoError(t, err)
	require.Equal(t, "please continue by filling out the interactive form", response.Text)

	state := dialog.InstallDialogState{
		AppID:  apps.AppID(appID),
		TeamID: th.ServerTestHelper.BasicTeam.Id,
	}

	b, err := json.Marshal(state)
	require.NoError(t, err)

	request := model.SubmitDialogRequest{
		Type: "dialog_submission",
		Submission: map[string]interface{}{
			"concent": "notrequire",
			"secret":  "1234",
		},
		State: string(b),
	}

	th.TestForSystemAdmin(t, func(t *testing.T, client *mmclient.ClientPP) {
		dialogResponse := client.SubmitDialog(request)
		require.Equal(t, http.StatusOK, dialogResponse.StatusCode)

		bindings, err := client.GetBindings()
		require.NoError(t, err)
		require.Len(t, bindings, 3)
		require.Equal(t, apps.LocationCommand, bindings[0].Location)
		require.Equal(t, apps.LocationChannelHeader, bindings[1].Location)
		require.Equal(t, apps.LocationPostMenu, bindings[2].Location)

		call := &apps.CallRequest{
			Call: apps.Call{
				Path: "/oks/ok/bla",
			},
			Context: &apps.Context{
				AppID:     apps.AppID(appID),
				Location:  apps.LocationCommand,
				UserID:    th.ServerTestHelper.App.Session().UserId,
				ChannelID: channelID,
			},
		}
		callResponse, err := client.Call(call)
		require.NoError(t, err)
		require.Equal(t, "OK", string(callResponse.Markdown))
	})
}
