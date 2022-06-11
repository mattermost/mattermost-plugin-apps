// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"encoding/json"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const echoID = apps.AppID("echotest")

func Echo(creq goapp.CallRequest) apps.CallResponse {
	return apps.NewTextResponse(utils.ToJSON(creq))
}

func newEchoApp() *goapp.App {
	var echoBindable = goapp.MakeBindableFormOrPanic(
		"echo",
		apps.Form{
			Icon:   "icon.png",
			Fields: []apps.Field{{Name: "test"}},
		},
		Echo,
	)

	return goapp.MakeAppOrPanic(
		apps.Manifest{
			AppID:       echoID,
			Version:     "v1.1.0",
			DisplayName: "Echos call requests as text/json",
			Icon:        "icon.png",
			HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/test/restapitest",
		},
		goapp.WithStatic(static),
		goapp.WithCommand(echoBindable),
		goapp.WithPostMenu(echoBindable),
		goapp.WithChannelHeader(echoBindable),
	)
}

func testEcho(th *Helper) {
	th.InstallAppWithCleanup(newEchoApp())

	th.Run("simple", func(th *Helper) {
		cresp := th.HappyCall(echoID, apps.CallRequest{
			Call: *apps.NewCall("/echo"),
			Values: model.StringInterface{
				"name": "value",
			},
		})

		echoResp := apps.CallRequest{}
		err := json.Unmarshal([]byte(cresp.Text), &echoResp)
		require.NoError(th, err)

		require.NotEmpty(th, echoResp.Context.ExpandedContext.BotUserID)
		echoResp.Context.ExpandedContext.BotUserID = ""
		require.NotEmpty(th, echoResp.Context.ExpandedContext.BotAccessToken)
		echoResp.Context.ExpandedContext.BotAccessToken = ""
		require.NotEmpty(th, echoResp.Context.ExpandedContext.MattermostSiteURL)
		echoResp.Context.ExpandedContext.MattermostSiteURL = ""

		require.EqualValues(th, apps.CallRequest{
			Call: apps.Call{
				Path: "/echo",
			},
			Values: map[string]interface{}{
				"name": "value",
			},
			Context: apps.Context{
				ExpandedContext: apps.ExpandedContext{
					AppPath:       "/plugins/com.mattermost.apps/apps/" + string(echoID),
					DeveloperMode: true,
				},
			},
		}, echoResp)
	})

	th.Run("AppsMetadata in response", func(th *Helper) {
		proxyResponse, _, err := th.CallWithAppMetadata(echoID, apps.CallRequest{
			Call: *apps.NewCall("/echo"),
		})

		require.NoError(th, err)
		require.Equal(th, string(echoID), proxyResponse.AppMetadata.BotUsername)
		require.NotEmpty(th, proxyResponse.AppMetadata.BotUserID)
	})
}
