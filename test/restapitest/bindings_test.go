// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	_ "embed"
	"encoding/json"
	"fmt"
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
		app.creq = apps.CallRequest{}
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
				// Top-level bindings are sorted by Location.
				expectedBindings: []apps.Binding{expectedChannelHeaderBinding, expectedCommandBinding, expectedPostMenuBinding},
			},
			"no locations": {
				requestedLocations: apps.Locations{},
				expectedBindings:   []apps.Binding{},
				expectedError:      "1 error occurred:\n\t* no location granted to bind to: forbidden\n\n",
			},
			"command only": {
				requestedLocations: apps.Locations{apps.LocationCommand},
				expectedBindings:   []apps.Binding{expectedCommandBinding},
				expectedError:      "2 errors occurred:\n\t* /channel_header: location is not granted: forbidden\n\t* /post_menu: location is not granted: forbidden\n\n",
			},
			"channel header only": {
				requestedLocations: apps.Locations{apps.LocationChannelHeader},
				expectedBindings:   []apps.Binding{expectedChannelHeaderBinding},
				expectedError:      "2 errors occurred:\n\t* /command: location is not granted: forbidden\n\t* /post_menu: location is not granted: forbidden\n\n",
			},
			"post menu only": {
				requestedLocations: apps.Locations{apps.LocationPostMenu},
				expectedBindings:   []apps.Binding{expectedPostMenuBinding},
				expectedError:      "2 errors occurred:\n\t* /command: location is not granted: forbidden\n\t* /channel_header: location is not granted: forbidden\n\n",
			},
			"command and post menu": {
				requestedLocations: apps.Locations{apps.LocationCommand, apps.LocationPostMenu},
				expectedBindings:   []apps.Binding{expectedCommandBinding, expectedPostMenuBinding},
				expectedError:      "1 error occurred:\n\t* /channel_header: location is not granted: forbidden\n\n",
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

				app.creq = apps.CallRequest{}
				out, err := httpGetBindings(th, th.ServerTestHelper.BasicChannel.TeamId, th.ServerTestHelper.BasicChannel.Id)
				require.NoError(err)
				require.Equal(tc.expectedError, out.Err)
				require.EqualValues(tc.expectedBindings, out.Bindings)

				if len(tc.requestedLocations) == 0 {
					th.Run("app with no locations doesn't get a request", func(*Helper) {
						require.Empty(app.creq)
					})
				}
			})
		}
	})

	th.Run("disabled app doesn't get a request", func(th *Helper) {
		appID := apps.AppID("disabled-app")

		appBindings := []apps.Binding{
			{
				Location: "loc-test",
				Label:    "lab-test",
				Submit:   apps.NewCall("/does-not-matter"),
			},
		}
		app := newBindingsApp(th, appID, nil,
			[]apps.Binding{
				{Location: apps.LocationCommand, Bindings: appBindings},
				{Location: apps.LocationChannelHeader, Bindings: appBindings},
				{Location: apps.LocationPostMenu, Bindings: appBindings},
			}).
			WithLocations(apps.Locations{apps.LocationCommand, apps.LocationChannelHeader, apps.LocationPostMenu}).
			WithPermissions(apps.Permissions{apps.PermissionActAsUser})

		th.InstallAppWithCleanup(app.App)
		th.DisableApp(app.App)
		require := require.New(th)

		app.creq = apps.CallRequest{}
		out, err := httpGetBindings(th, th.ServerTestHelper.BasicChannel.TeamId, th.ServerTestHelper.BasicChannel.Id)
		require.NoError(err)
		require.Equal("1 error occurred:\n\t* app is disabled by the administrator: disabled-app: forbidden\n\n", out.Err)
		require.Empty(app.creq)
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

		appsURL := fmt.Sprintf("http://localhost:%v/plugins/com.mattermost.apps/apps", th.ServerTestHelper.Server.ListenAddr.Port)
		expected := []apps.Binding{
			{
				Location: "/command",
				Bindings: []apps.Binding{
					{
						AppID:       "bind1",
						Location:    "baseCommandLocation",
						Icon:        appsURL + "/bind1/static/base command icon",
						Label:       "baseCommandLabel",
						Hint:        "base command hint",
						Description: "base command description",
						Bindings: []apps.Binding{
							{
								AppID:       "bind1",
								Location:    "message",
								Icon:        "https://example.com/image.png",
								Label:       "message",
								Hint:        "message command hint",
								Description: "message command description",
								Submit:      apps.NewCall("/path"),
							},
							{
								AppID:       "bind1",
								Location:    "message-modal",
								Icon:        appsURL + "/bind1/static/message-modal command icon",
								Label:       "message-modal",
								Hint:        "message-modal command hint",
								Description: "message-modal command description",
								Submit:      apps.NewCall("/path"),
							},
							{
								AppID:       "bind1",
								Location:    "manage",
								Icon:        appsURL + "/bind1/static/valid/path",
								Label:       "manage",
								Hint:        "manage command hint",
								Description: "manage command description",
								Bindings: []apps.Binding{
									{
										AppID:       "bind1",
										Location:    "subscribe",
										Icon:        appsURL + "/bind1/static/subscribe command icon",
										Label:       "subscribe",
										Hint:        "subscribe command hint",
										Description: "subscribe command description",
										Submit:      apps.NewCall("/path"),
									},
									{
										AppID:       "bind1",
										Location:    "unsubscribe",
										Icon:        appsURL + "/bind1/static/unsubscribe command icon",
										Label:       "unsubscribe",
										Hint:        "unsubscribe command hint",
										Description: "unsubscribe command description",
										Submit:      apps.NewCall("/path"),
									},
								},
							},
						},
					},
					{
						AppID:       "bind2",
						Location:    "app2BaseCommandLocation",
						Icon:        appsURL + "/bind2/static/app2-base-command-icon",
						Label:       "app2BaseCommandLabel",
						Hint:        "app2 base command hint",
						Description: "app2 base command description",
						Bindings: []apps.Binding{
							{
								AppID:       "bind2",
								Location:    "connect",
								Icon:        appsURL + "/bind2/static/connect command icon",
								Label:       "connect",
								Hint:        "connect command hint",
								Description: "connect command description",
								Submit:      apps.NewCall("/path"),
							},
						},
					},
				},
			},
		}

		require.EqualValues(expected, out.Bindings)
	})
}
