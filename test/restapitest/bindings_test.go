// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	_ "embed"
	"encoding/json"
	"io"
	"net/url"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
	appspath "github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type bindingsApp struct {
	*goapp.App
	creq  apps.CallRequest
	cresp apps.CallResponse
}

//go:embed testdata/bindings_multiple_apps_have_commands.json
var multipleAppsHaveCommandsJSON []byte

func newBindingsApp(th *Helper, appID apps.AppID, bindExpand *apps.Expand, bindings []apps.Binding) *bindingsApp {
	app := bindingsApp{}
	app.App = goapp.MakeAppOrPanic(
		apps.Manifest{
			AppID:       appID,
			Version:     "v1.0.0",
			DisplayName: "Returns bindings",
			HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/test/restapitest",
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

func (a *bindingsApp) WithLocations(requested apps.Locations) *bindingsApp {
	a.Manifest.RequestedLocations = requested
	return a
}

func (a *bindingsApp) WithPermissions(requested apps.Permissions) *bindingsApp {
	a.Manifest.RequestedPermissions = requested
	return a
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
			if len(b.Bindings) == 0 {
				continue
			}
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
		th.Skip()
		appID := apps.AppID("context_bindings")
		app := newBindingsApp(th, appID,
			// Expand everything we realistically can.
			&apps.Expand{
				ActingUser:            apps.ExpandSummary.Required(),
				ActingUserAccessToken: apps.ExpandAll.Required(),
				Channel:               apps.ExpandSummary.Required(),
				Team:                  apps.ExpandSummary.Required(),
				ChannelMember:         apps.ExpandSummary.Required(),
				TeamMember:            apps.ExpandSummary.Required(),
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
			}).
			WithLocations(apps.Locations{apps.LocationCommand, apps.LocationChannelHeader, apps.LocationPostMenu}).
			WithPermissions(apps.Permissions{apps.PermissionActAsUser})

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

	th.Run("bindings are accepted only for requested locations", func(th *Helper) {
		appID := apps.AppID("location_bindings")

		appBindings := []apps.Binding{
			{
				Location: "loc-test",
				Label:    "lab-test",
				Submit:   apps.NewCall("/does-not-matter"),
			},
		}

		expectedCommandBinding := apps.Binding{
			Location: apps.LocationCommand,
			Bindings: []apps.Binding{
				{
					AppID:    appID,
					Label:    "lab-test",
					Location: "loc-test",
					Submit:   apps.NewCall("/does-not-matter"),
				},
			},
		}

		expectedChannelHeaderBinding := apps.Binding{
			Location: apps.LocationChannelHeader,
			Bindings: []apps.Binding{
				{
					AppID:    appID,
					Label:    "lab-test",
					Location: "loc-test",
					Submit:   apps.NewCall("/does-not-matter"),
				},
			},
		}

		expectedPostMenuBinding := apps.Binding{
			Location: apps.LocationPostMenu,
			Bindings: []apps.Binding{
				{
					AppID:    appID,
					Label:    "lab-test",
					Location: "loc-test",
					Submit:   apps.NewCall("/does-not-matter"),
				},
			},
		}

		for name, tc := range map[string]struct {
			requestedLocations apps.Locations
			expectedBindings   []apps.Binding
			expectedError      string
		}{
			"all locations": {
				requestedLocations: apps.Locations{apps.LocationCommand, apps.LocationChannelHeader, apps.LocationPostMenu},
				expectedBindings:   []apps.Binding{expectedCommandBinding, expectedChannelHeaderBinding, expectedPostMenuBinding},
			},
			"no locations": {
				requestedLocations: apps.Locations{},
				expectedBindings:   []apps.Binding{},
				expectedError:      "1 error occurred:\n\t* location_bindings: no location granted to bind to: forbidden\n\n",
			},
			"command only": {
				requestedLocations: apps.Locations{apps.LocationCommand},
				expectedBindings:   []apps.Binding{expectedCommandBinding},
				expectedError:      "1 error occurred:\n\t* location_bindings: 2 errors occurred:\n\t* /channel_header: location is not granted: forbidden\n\t* /post_menu: location is not granted: forbidden\n\n\n\n",
			},
			"channel header only": {
				requestedLocations: apps.Locations{apps.LocationChannelHeader},
				expectedBindings:   []apps.Binding{expectedChannelHeaderBinding},
				expectedError:      "1 error occurred:\n\t* location_bindings: 2 errors occurred:\n\t* /command: location is not granted: forbidden\n\t* /post_menu: location is not granted: forbidden\n\n\n\n",
			},
			"post menu only": {
				requestedLocations: apps.Locations{apps.LocationPostMenu},
				expectedBindings:   []apps.Binding{expectedPostMenuBinding},
				expectedError:      "1 error occurred:\n\t* location_bindings: 2 errors occurred:\n\t* /command: location is not granted: forbidden\n\t* /channel_header: location is not granted: forbidden\n\n\n\n",
			},
			"command and post menu": {
				requestedLocations: apps.Locations{apps.LocationCommand, apps.LocationPostMenu},
				expectedBindings:   []apps.Binding{expectedCommandBinding, expectedPostMenuBinding},
				expectedError:      "1 error occurred:\n\t* location_bindings: 1 error occurred:\n\t* /channel_header: location is not granted: forbidden\n\n\n\n",
			},
		} {
			th.Run(name, func(th *Helper) {
				app := newBindingsApp(th, appID, nil,
					[]apps.Binding{
						{Location: apps.LocationCommand, Bindings: appBindings},
						{Location: apps.LocationChannelHeader, Bindings: appBindings},
						{Location: apps.LocationPostMenu, Bindings: appBindings},
					}).
					WithLocations(tc.requestedLocations).
					WithPermissions(apps.Permissions{apps.PermissionActAsUser})

				th.InstallAppWithCleanup(app.App)
				require := require.New(th)

				out, err := httpGetBindings(th, th.ServerTestHelper.BasicChannel.TeamId, th.ServerTestHelper.BasicChannel.Id)
				require.NoError(err)
				require.Equal(tc.expectedError, out.Err)
				// require.Equal("", utils.Pretty(app.cresp.Data))
				require.Equal(utils.Pretty(tc.expectedBindings), utils.Pretty(out.Bindings))
			})
		}
	})

	th.Run("multiple apps have commands", func(th *Helper) {
		app1ID := apps.AppID("bind1")
		app1 := newBindingsApp(th, app1ID, nil,
			[]apps.Binding{
				{
					Location: apps.LocationCommand,
					Bindings: []apps.Binding{
						{
							Location:    "baseCommandLocation",
							Label:       "baseCommandLabel",
							Icon:        "base command icon",
							Hint:        "base command hint",
							Description: "base command description",
							Bindings: []apps.Binding{
								{
									Location:    "message",
									Label:       "message",
									Icon:        "https://example.com/image.png",
									Hint:        "message command hint",
									Description: "message command description",
									Submit:      &apps.Call{Path: "/path"},
								}, {
									Location:    "message-modal",
									Label:       "message-modal",
									Icon:        "message-modal command icon",
									Hint:        "message-modal command hint",
									Description: "message-modal command description",
									Submit:      &apps.Call{Path: "/path"},
								}, {
									Location:    "manage",
									Label:       "manage",
									Icon:        "valid/path",
									Hint:        "manage command hint",
									Description: "manage command description",
									Bindings: []apps.Binding{
										{
											Location:    "subscribe",
											Label:       "subscribe",
											Icon:        "subscribe command icon",
											Hint:        "subscribe command hint",
											Description: "subscribe command description",
											Submit:      &apps.Call{Path: "/path"},
										}, {
											Location:    "unsubscribe",
											Label:       "unsubscribe",
											Icon:        "unsubscribe command icon",
											Hint:        "unsubscribe command hint",
											Description: "unsubscribe command description",
											Submit:      &apps.Call{Path: "/path"},
										},
									},
								},
							},
						},
					},
				},
			}).
			WithLocations(apps.Locations{apps.LocationCommand})

		app2ID := apps.AppID("bind2")
		app2 := newBindingsApp(th, app2ID, nil,
			[]apps.Binding{
				{
					Location: apps.LocationCommand,
					Bindings: []apps.Binding{
						{
							Location:    "app2BaseCommandLocation",
							Label:       "app2BaseCommandLabel",
							Icon:        "app2-base-command-icon",
							Hint:        "app2 base command hint",
							Description: "app2 base command description",
							Bindings: []apps.Binding{
								{
									Location:    "connect",
									Label:       "connect",
									Icon:        "connect command icon",
									Hint:        "connect command hint",
									Description: "connect command description",
									Submit:      &apps.Call{Path: "/path"},
								},
							},
						},
					},
				},
			}).
			WithLocations(apps.Locations{apps.LocationCommand})

		require := require.New(th)
		th.InstallAppWithCleanup(app1.App)
		th.InstallAppWithCleanup(app2.App)

		out, err := httpGetBindings(th, "", "")
		require.NoError(err)
		require.Equal("", out.Err)
		th.EqualBindings([]byte(multipleAppsHaveCommandsJSON), out.Bindings)
	})
}
