// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"encoding/json"
	"io"
	"net/url"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
	appspath "github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

const contextBindingsID = apps.AppID("contextbindingstest")

type contextBindingsApp struct {
	*goapp.App
	in  apps.CallRequest
	out apps.CallResponse
}

var contextBindingsExpand = apps.Expand{
	ActingUser:            apps.ExpandSummary.Required(),
	ActingUserAccessToken: apps.ExpandAll.Required(),
	Channel:               apps.ExpandSummary.Required(),
	Team:                  apps.ExpandSummary.Optional(),
	Post:                  apps.ExpandSummary.Optional(),
}

func newContextBindingsApp(th *Helper) *contextBindingsApp {
	app := contextBindingsApp{}
	app.App = goapp.MakeAppOrPanic(
		apps.Manifest{
			AppID:                contextBindingsID,
			Version:              "v1.0.0",
			DisplayName:          "Returns context-specific bindings",
			HomepageURL:          "https://github.com/mattermost/mattermost-plugin-apps/test/restapitest",
			RequestedLocations:   apps.Locations{apps.LocationCommand, apps.LocationChannelHeader, apps.LocationPostMenu},
			RequestedPermissions: apps.Permissions{apps.PermissionActAsUser},
		},
		goapp.WithBindingsExpand(contextBindingsExpand),
		goapp.TestWithBindingsHandler(
			func(creq goapp.CallRequest) apps.CallResponse {
				app.in = creq.CallRequest
				commandBinding := apps.Binding{
					Location: apps.LocationCommand,
					Bindings: []apps.Binding{
						{
							Location: apps.Location(contextBindingsID),
							Bindings: []apps.Binding{
								{
									Location: apps.Location("test"),
									Submit:   apps.NewCall("/echo"),
								},
							},
						},
					},
				}
				cresp := apps.NewDataResponse([]apps.Binding{commandBinding})
				app.out = cresp
				return cresp
			},
		),
	)

	app.HandleCall("/echo", Echo)
	app.HandleCall("/last/in", func(_ goapp.CallRequest) apps.CallResponse {
		return apps.NewDataResponse(app.in)
	})
	app.HandleCall("/last/out", func(_ goapp.CallRequest) apps.CallResponse {
		return apps.NewDataResponse(app.out)
	})

	return &app
}

func testContextBindings(th *Helper) {
	app := newContextBindingsApp(th)
	th.InstallAppWithCleanup(app.App)

	th.Run("context-specific bindings", func(th *Helper) {
		require := require.New(th)

		q := url.Values{}
		q.Set(config.PropChannelID, th.ServerTestHelper.BasicChannel.Id)
		q.Set(config.PropTeamID, th.ServerTestHelper.BasicTeam.Id)
		q.Set("test", "true")

		url := th.UserClientPP.GetPluginRoute(appclient.AppsPluginName) + appspath.API + "/bindings?" + q.Encode()
		httpResp, err := th.UserClientPP.DoAPIGET(url, "")
		require.NoError(err)
		require.NotEmpty(httpResp)
		defer httpResp.Body.Close()
		data, err := io.ReadAll(httpResp.Body)
		require.NoError(err)
		out := struct {
			Bindings []apps.Binding `json:"bindings"`
			Err      string         `json:"error"`
		}{}
		err = json.Unmarshal(data, &out)
		require.NoError(err)
		require.Equal("", out.Err)
		require.NotEmpty(out.Bindings)
		require.Equal(apps.LocationCommand, out.Bindings[0].Location)
		var mine *apps.Binding
		for _, b := range out.Bindings[0].Bindings {
			if b.AppID == contextBindingsID {
				clone := b
				mine = &clone
				break
			}
		}
		require.NotEmpty(mine)
		require.EqualValues(&apps.Binding{
			AppID:    contextBindingsID,
			Location: apps.Location(contextBindingsID),
			Label:    string(contextBindingsID),
			Bindings: []apps.Binding{
				{
					AppID:    contextBindingsID,
					Label:    "test",
					Location: apps.Location("test"),
					Submit:   apps.NewCall("/echo"),
				},
			},
		}, mine)

		require.NotEmpty(app.in)
		require.EqualValues(th.ServerTestHelper.BasicChannel.TeamId, app.in.Context.ExpandedContext.Team.Id)
		require.EqualValues(th.ServerTestHelper.BasicChannel.Id, app.in.Context.ExpandedContext.Channel.Id)
	})
}
