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

type bindingsApp struct {
	*goapp.App
	creq  apps.CallRequest
	cresp apps.CallResponse
}

func newBindingsApp(th *Helper, appID apps.AppID, bindExpand *apps.Expand, bindings []apps.Binding) *bindingsApp {
	app := bindingsApp{}
	app.App = goapp.MakeAppOrPanic(
		apps.Manifest{
			AppID:                appID,
			Version:              "v1.0.0",
			DisplayName:          "Returns bindings",
			HomepageURL:          "https://github.com/mattermost/mattermost-plugin-apps/test/restapitest",
			RequestedLocations:   apps.Locations{apps.LocationCommand, apps.LocationChannelHeader, apps.LocationPostMenu},
			RequestedPermissions: apps.Permissions{apps.PermissionActAsUser},
		},
		goapp.WithBindingsExpand(bindExpand),
		goapp.TestWithBindingsHandler(
			func(creq goapp.CallRequest) apps.CallResponse {
				app.creq = creq.CallRequest
				app.cresp = apps.NewDataResponse(bindings)
				return app.cresp
			},
		),
	)

	app.HandleCall("/echo", Echo)
	app.HandleCall("/creq", func(_ goapp.CallRequest) apps.CallResponse {
		return apps.NewDataResponse(app.creq)
	})
	app.HandleCall("/cresp", func(_ goapp.CallRequest) apps.CallResponse {
		return apps.NewDataResponse(app.cresp)
	})

	return &app
}

func removeAppsPluginBindings(in []apps.Binding) []apps.Binding {
	// remove the apps plugin's own bindings
	filtered := []apps.Binding{}
	for _, b := range in {
		if b.AppID == apps.AppID("apps") {
			continue
		}
		if len(b.Bindings) != 0 {
			b.Bindings = removeAppsPluginBindings(b.Bindings)
		}
		filtered = append(filtered, b)
	}
	return filtered
}

func testBindings(th *Helper) {
	// httpGetBindings makes an HTTP GET /bindings request to the proxy
	type getBindingsOut struct {
		Bindings []apps.Binding `json:"bindings"`
		Err      string         `json:"error"`
	}
	httpGetBindings := func(th *Helper, teamID, channelID string) (getBindingsOut, error) {
		require := require.New(th)

		// Set the HTTP request query args. ?test=true makes the REST API return
		// the errors in addition to the bindings
		q := url.Values{}
		q.Set(config.PropTeamID, teamID)
		q.Set(config.PropChannelID, channelID)
		q.Set(config.PropUserAgent, "test-agent")
		q.Set("test", "true")

		url := th.UserClientPP.GetPluginRoute(appclient.AppsPluginName) + appspath.API + "/bindings?" + q.Encode()
		httpResp, err := th.UserClientPP.DoAPIGET(url, "")
		require.NoError(err)
		require.NotEmpty(httpResp)
		defer httpResp.Body.Close()
		data, err := io.ReadAll(httpResp.Body)
		require.NoError(err)
		// the output of the /bindings test mode
		out := getBindingsOut{}
		err = json.Unmarshal(data, &out)
		out.Bindings = removeAppsPluginBindings(out.Bindings)
		return out, err
	}

	th.Run("expand all HTTP GET query args", func(th *Helper) {
		appID := apps.AppID("context_bindings")
		app := newBindingsApp(th, appID,
			&apps.Expand{
				ActingUser:            apps.ExpandSummary.Required(),
				ActingUserAccessToken: apps.ExpandAll.Required(),
				Channel:               apps.ExpandSummary.Required(),
				Team:                  apps.ExpandSummary.Optional(),
				Post:                  apps.ExpandSummary.Optional(),
			},
			[]apps.Binding{
				{
					Location: apps.LocationCommand,
					Bindings: []apps.Binding{
						{
							Location: "top-level",
							Bindings: []apps.Binding{
								{
									Location: "subcommand",
									Submit:   apps.NewCall("/does-not-matter"),
								},
							},
						},
					},
				},
			},
		)
		th.InstallAppWithCleanup(app.App)
		require := require.New(th)

		require.Equal(th.ServerTestHelper.BasicPost.ChannelId, th.ServerTestHelper.BasicChannel.Id)
		out, err := httpGetBindings(th, th.ServerTestHelper.BasicChannel.TeamId, th.ServerTestHelper.BasicChannel.Id)
		require.NoError(err)
		require.Equal("", out.Err)
		require.EqualValues([]apps.Binding{
			{
				Location: apps.LocationCommand,
				Bindings: []apps.Binding{
					{
						AppID:    appID,
						Location: "top-level",
						Label:    "top-level",
						Bindings: []apps.Binding{
							{
								// TODO <>/<> ticket: eliminate appID on sub-bindings if not TLB.
								AppID:    appID,
								Location: "subcommand",
								Label:    "subcommand",
								Submit:   apps.NewCall("/does-not-matter"),
							},
						},
					},
				},
			},
		}, out.Bindings)

		require.NotEmpty(app.creq)
		require.EqualValues(th.ServerTestHelper.BasicChannel.TeamId, app.creq.Context.ExpandedContext.Team.Id)
		require.EqualValues(th.ServerTestHelper.BasicChannel.Id, app.creq.Context.ExpandedContext.Channel.Id)
	})
}
