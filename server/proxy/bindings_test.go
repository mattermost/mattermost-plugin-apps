package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_upstream"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func testBinding(appID apps.AppID, parent apps.Location, n string) []apps.Binding {
	return []apps.Binding{
		{
			AppID:    appID,
			Location: parent,
			Bindings: []apps.Binding{
				{
					AppID:    appID,
					Location: apps.Location(fmt.Sprintf("id-%s", n)),
					Hint:     fmt.Sprintf("hint-%s", n),
				},
			},
		},
	}
}

func TestMergeBindings(t *testing.T) {
	type TC struct {
		name               string
		bb1, bb2, expected []apps.Binding
	}

	for _, tc := range []TC{
		{
			name: "happy simplest",
			bb1: []apps.Binding{
				{
					Location: "1",
				},
			},
			bb2: []apps.Binding{
				{
					Location: "2",
				},
			},
			expected: []apps.Binding{
				{
					Location: "1",
				},
				{
					Location: "2",
				},
			},
		},
		{
			name:     "happy simple 1",
			bb1:      testBinding("app1", apps.LocationCommand, "simple"),
			bb2:      nil,
			expected: testBinding("app1", apps.LocationCommand, "simple"),
		},
		{
			name:     "happy simple 2",
			bb1:      nil,
			bb2:      testBinding("app1", apps.LocationCommand, "simple"),
			expected: testBinding("app1", apps.LocationCommand, "simple"),
		},
		{
			name:     "happy simple same",
			bb1:      testBinding("app1", apps.LocationCommand, "simple"),
			bb2:      testBinding("app1", apps.LocationCommand, "simple"),
			expected: testBinding("app1", apps.LocationCommand, "simple"),
		},
		{
			name: "happy simple merge",
			bb1:  testBinding("app1", apps.LocationPostMenu, "simple"),
			bb2:  testBinding("app1", apps.LocationCommand, "simple"),
			expected: append(
				testBinding("app1", apps.LocationPostMenu, "simple"),
				testBinding("app1", apps.LocationCommand, "simple")...,
			),
		},
		{
			name: "happy simple 2 apps",
			bb1:  testBinding("app1", apps.LocationCommand, "simple"),
			bb2:  testBinding("app2", apps.LocationCommand, "simple"),
			expected: append(
				testBinding("app1", apps.LocationCommand, "simple"),
				testBinding("app2", apps.LocationCommand, "simple")...,
			),
		},
		{
			name: "happy 2 simple commands",
			bb1:  testBinding("app1", apps.LocationCommand, "simple1"),
			bb2:  testBinding("app1", apps.LocationCommand, "simple2"),
			expected: []apps.Binding{
				{
					AppID:    "app1",
					Location: "/command",
					Bindings: []apps.Binding{
						{
							AppID:    "app1",
							Location: "id-simple1",
							Hint:     "hint-simple1",
						},
						{
							AppID:    "app1",
							Location: "id-simple2",
							Hint:     "hint-simple2",
						},
					},
				},
			},
		},
		{
			name: "happy 2 apps",
			bb1: []apps.Binding{
				{
					Location: "/post_menu",
					Bindings: []apps.Binding{
						{
							AppID:       "zendesk",
							Label:       "Create zendesk ticket",
							Description: "Create ticket in zendesk",
							Form: &apps.Form{
								Submit: apps.NewCall("http://localhost:4000/create"),
							},
						},
					},
				},
			},
			bb2: []apps.Binding{
				{
					Location: "/post_menu",
					Bindings: []apps.Binding{
						{
							AppID:       "hello",
							Label:       "Create hello ticket",
							Description: "Create ticket in hello",
							Form: &apps.Form{
								Submit: apps.NewCall("http://localhost:4000/hello"),
							},
						},
					},
				},
			},
			expected: []apps.Binding{
				{
					Location: "/post_menu",
					Bindings: []apps.Binding{
						{
							AppID:       "zendesk",
							Label:       "Create zendesk ticket",
							Description: "Create ticket in zendesk",
							Form: &apps.Form{
								Submit: apps.NewCall("http://localhost:4000/create"),
							},
						},
						{
							AppID:       "hello",
							Label:       "Create hello ticket",
							Description: "Create ticket in hello",
							Form: &apps.Form{
								Submit: apps.NewCall("http://localhost:4000/hello"),
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := mergeBindings(tc.bb1, tc.bb2)
			equalBindings(t, tc.expected, out)
		})
	}
}

// equalBindings asserts that two slices of bindings are equal ignoring the order of the elements.
// If there are duplicate elements, the number of appearances of each of them in both lists should match.
//
// equalBindings calls t.Fail if the elements not match.
func equalBindings(t *testing.T, expected, actual []apps.Binding) {
	opt := cmpopts.SortSlices(func(a apps.Binding, b apps.Binding) bool {
		return a.AppID < b.AppID
	})

	if diff := cmp.Diff(expected, actual, opt); diff != "" {
		t.Errorf("Bindings mismatch (-expected +actual):\n%s", diff)
	}
}

func TestCleanAppBinding(t *testing.T) {
	app := &apps.App{
		Manifest: apps.Manifest{
			AppID: "appid",
		},
		GrantedLocations: apps.Locations{
			apps.LocationCommand,
			apps.LocationChannelHeader,
		},
	}

	type TC struct {
		in               apps.Binding
		locPrefix        apps.Location
		userAgent        string
		expected         *apps.Binding
		expectedProblems string
	}

	for name, tc := range map[string]TC{
		"happy simplest": {
			in: apps.Binding{
				Location: "test",
				Submit:   apps.NewCall("/hello"),
			},
			locPrefix: apps.LocationCommand.Sub("main-command"),
			expected: &apps.Binding{
				AppID:    "appid",
				Location: "test",
				Label:    "test",
				Submit:   apps.NewCall("/hello"),
			},
		},
		"trim location": {
			in: apps.Binding{
				Location: " test-1 \t",
				Submit:   apps.NewCall("/hello"),
			},
			locPrefix: apps.LocationCommand.Sub("main-command"),
			expected: &apps.Binding{
				AppID:    "appid",
				Location: "test-1",
				Label:    "test-1",
				Submit:   apps.NewCall("/hello"),
			},
			expectedProblems: "1 error occurred:\n\t* /command/main-command/test-1: trimmed whitespace from location\n\n",
		},
		"ERROR location PostMenu not granted": {
			in: apps.Binding{
				Location: "test",
				Submit:   apps.NewCall("/hello"),
			},
			locPrefix:        apps.LocationPostMenu,
			expected:         nil,
			expectedProblems: "1 error occurred:\n\t* /post_menu/test: location is not granted: forbidden\n\n",
		},
		"trim command label": {
			in: apps.Binding{
				Location: "test",
				Label:    "\ntest-label \t",
				Submit:   apps.NewCall("/hello"),
			},
			locPrefix: apps.LocationCommand.Sub("main-command"),
			expected: &apps.Binding{
				AppID:    "appid",
				Location: "test",
				Label:    "test-label",
				Submit:   apps.NewCall("/hello"),
			},
			expectedProblems: "1 error occurred:\n\t* /command/main-command/test: trimmed whitespace from label test-label\n\n",
		},
		"label defaults to location for command": {
			in: apps.Binding{
				Location: "test",
				Submit:   apps.NewCall("/hello"),
			},
			locPrefix: apps.LocationCommand.Sub("main-command"),
			expected: &apps.Binding{
				AppID:    "appid",
				Location: "test",
				Label:    "test",
				Submit:   apps.NewCall("/hello"),
			},
		},
		"label does not default for non-commands": {
			in: apps.Binding{
				Location: "test",
				Submit:   apps.NewCall("/hello"),
			},
			locPrefix: apps.LocationChannelHeader.Sub("some"),
			expected: &apps.Binding{
				AppID:    "appid",
				Location: "test",
				Submit:   apps.NewCall("/hello"),
			},
		},
		"ERROR neither location nor label": {
			in: apps.Binding{
				Submit: apps.NewCall("/hello"),
			},
			locPrefix:        apps.LocationCommand.Sub("main-command"),
			expected:         nil,
			expectedProblems: "1 error occurred:\n\t* /command/main-command: sub-binding with no location nor label\n\n",
		},
		"ERROR whitsepace in command label": {
			in: apps.Binding{
				Location: "test",
				Label:    "test label",
				Submit:   apps.NewCall("/hello"),
			},
			locPrefix:        apps.LocationCommand.Sub("main-command"),
			expected:         nil,
			expectedProblems: "1 error occurred:\n\t* /command/main-command/test: command label \"test label\" has multiple words\n\n",
		},
		"normalize icon path": {
			in: apps.Binding{
				Location: "test",
				Submit:   apps.NewCall("/hello"),
				Icon:     "a///static.icon",
			},
			locPrefix: apps.LocationCommand.Sub("main-command"),
			expected: &apps.Binding{
				AppID:    "appid",
				Location: "test",
				Label:    "test",
				Icon:     "/apps/appid/static/a/static.icon",
				Submit:   apps.NewCall("/hello"),
			},
		},
		"invalid icon path": {
			in: apps.Binding{
				Submit:   apps.NewCall("/hello"),
				Location: "test",
				Icon:     "../a/...//static.icon",
			},
			locPrefix: apps.LocationCommand.Sub("main-command"),
			expected: &apps.Binding{
				AppID:    "appid",
				Location: "test",
				Label:    "test",
				Submit:   apps.NewCall("/hello"),
			},
			expectedProblems: "1 error occurred:\n\t* /command/main-command/test: invalid icon path \"../a/...//static.icon\" in binding\n\n",
		},
		"ERROR: icon required for ChannelHeader in webapp": {
			in: apps.Binding{
				Location: "test",
				Submit:   apps.NewCall("/hello"),
			},
			locPrefix:        apps.LocationChannelHeader,
			userAgent:        "webapp",
			expected:         nil,
			expectedProblems: "1 error occurred:\n\t* /channel_header/test: no icon in channel header binding\n\n",
		},
		"icon not required for ChannelHeader in mobile": {
			in: apps.Binding{
				Location: "test",
				Submit:   apps.NewCall("/hello"),
			},
			locPrefix: apps.LocationChannelHeader,
			userAgent: "something-else",
			expected: &apps.Binding{
				AppID:    "appid",
				Location: "test",
				Submit:   apps.NewCall("/hello"),
			},
		},
		"ERROR: no submit/form/bindings": {
			in: apps.Binding{
				Location: "test",
			},
			locPrefix:        apps.LocationChannelHeader,
			expected:         nil,
			expectedProblems: "1 error occurred:\n\t* /channel_header/test: (only) one of  \"submit\", \"form\", or \"bindings\" must be set in a binding\n\n",
		},
		"ERROR: submit and form": {
			in: apps.Binding{
				Location: "test",
				Submit:   apps.NewCall("/hello"),
				Form:     apps.NewBlankForm(apps.NewCall("/hello")),
			},
			locPrefix:        apps.LocationChannelHeader,
			expected:         nil,
			expectedProblems: "1 error occurred:\n\t* /channel_header/test: (only) one of  \"submit\", \"form\", or \"bindings\" must be set in a binding\n\n",
		},
		"ERROR: submit and bindings": {
			in: apps.Binding{
				Location: "test",
				Submit:   apps.NewCall("/hello"),
				Bindings: []apps.Binding{
					{
						Location: "test1",
					},
					{
						Location: "test2",
					},
				},
			},
			locPrefix:        apps.LocationChannelHeader,
			expected:         nil,
			expectedProblems: "1 error occurred:\n\t* /channel_header/test: (only) one of  \"submit\", \"form\", or \"bindings\" must be set in a binding\n\n",
		},
		"clean sub-bindings": {
			in: apps.Binding{
				Location: "test",
				Bindings: []apps.Binding{
					{
						Location: "test1",
						Submit:   apps.NewCall("/hello"),
					},
					{
						Location: "test2",
						Submit:   apps.NewCall("/hello"),
					},
				},
			},
			locPrefix: apps.LocationChannelHeader,
			expected: &apps.Binding{
				AppID:    "appid",
				Location: "test",
				Bindings: []apps.Binding{
					{
						AppID:    "appid",
						Location: "test1",
						Submit:   apps.NewCall("/hello"),
					},
					{
						AppID:    "appid",
						Location: "test2",
						Submit:   apps.NewCall("/hello"),
					},
				},
			},
		},
		"clean form": {
			in: apps.Binding{
				Location: "test",
				Form: &apps.Form{
					Submit: apps.NewCall("/hello"),
					Fields: []apps.Field{
						{Name: "in valid"},
					},
				},
			},
			locPrefix: apps.LocationChannelHeader,
			expected: &apps.Binding{
				AppID:    "appid",
				Location: "test",
				Form: &apps.Form{
					Submit: apps.NewCall("/hello"),
					Fields: []apps.Field{},
				},
			},
			expectedProblems: "1 error occurred:\n\t* field name must be a single word: \"in valid\"\n\n",
		},
	} {
		t.Run(name, func(t *testing.T) {
			b, err := cleanAppBinding(app, tc.in, tc.locPrefix, tc.userAgent, config.Config{})
			if tc.expectedProblems != "" {
				require.Error(t, err)
				require.Equal(t, tc.expectedProblems, err.Error())
			} else {
				require.NoError(t, err)
				require.EqualValues(t, tc.expected, b)
			}
		})
	}
}

func TestRefreshBindingsEventAfterCall(t *testing.T) {
	tApps := []apps.App{
		{
			BotUserID:   "botid",
			BotUsername: "botusername",
			DeployType:  apps.DeployBuiltin,
			Manifest: apps.Manifest{
				AppID:       apps.AppID("app1"),
				DisplayName: "App 1",
			},
		},
	}

	creq := apps.CallRequest{
		Context: apps.Context{
			UserAgentContext: apps.UserAgentContext{
				AppID:  "app1",
				UserID: "userid",
			},
			ExpandedContext: apps.ExpandedContext{
				ActingUser: &model.User{
					Id: "userid",
				},
				App: &tApps[0],
			},
		},
		Call: apps.Call{
			Expand: &apps.Expand{
				App:        apps.ExpandNone,
				ActingUser: apps.ExpandNone,
			},
			Path: "/",
		},
	}

	type TC struct {
		name             string
		applications     []apps.App
		callRequest      apps.CallRequest
		callResponse     apps.CallResponse
		checkExpectation func(api *plugintest.API)
	}

	makeBindingRequest := func(req apps.CallRequest) apps.CallRequest {
		req.Call.Path = path.Bindings
		return req
	}

	for _, tc := range []TC{
		{
			name:         "refresh bindings when flag is set and OK response",
			applications: tApps,
			callRequest:  creq,
			callResponse: apps.CallResponse{
				Type:            apps.CallResponseTypeOK,
				RefreshBindings: true,
			},
			checkExpectation: func(testApi *plugintest.API) {
				testApi.On("PublishWebSocketEvent", config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{UserId: "userid"}).Once()
			},
		},
		{
			name:         "don't refresh bindings when flag is set and Error response",
			applications: tApps,
			callRequest:  creq,
			callResponse: apps.CallResponse{
				Type:            apps.CallResponseTypeError,
				RefreshBindings: true,
			},
			checkExpectation: func(testApi *plugintest.API) {
			},
		},
		{
			name:         "don't refresh bindings when flag is not set",
			applications: tApps,
			callRequest:  creq,
			callResponse: apps.CallResponse{
				Type:            apps.CallResponseTypeOK,
				RefreshBindings: false,
			},
			checkExpectation: func(testApi *plugintest.API) {
			},
		},
		{
			name:         "don't refresh when binding response is handled",
			applications: tApps,
			callRequest:  makeBindingRequest(creq),
			callResponse: apps.CallResponse{
				Type:            apps.CallResponseTypeOK,
				RefreshBindings: true,
			},
			checkExpectation: func(testApi *plugintest.API) {
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testAPI := &plugintest.API{}
			testDriver := &plugintest.Driver{}
			ctrl := gomock.NewController(t)

			conf := config.NewTestConfigService(nil).WithMattermostConfig(model.Config{
				ServiceSettings: model.ServiceSettings{
					SiteURL: model.NewString("test.mattermost.com"),
				},
			}).WithMattermostAPI(pluginapi.NewClient(testAPI, testDriver))

			appstore := store.TestAppStore{}

			upstreams := map[apps.AppID]upstream.Upstream{}
			for i := range tc.applications {
				app := tc.applications[i]

				up := mock_upstream.NewMockUpstream(ctrl)

				// set up an empty OK call response
				b, _ := json.Marshal(tc.callResponse)
				reader := io.NopCloser(bytes.NewReader(b))
				up.EXPECT().Roundtrip(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(reader, nil)

				upstreams[app.Manifest.AppID] = up
				_ = appstore.Save(nil, app)
			}

			proxy := &Proxy{
				appStore:         appstore,
				builtinUpstreams: upstreams,
				conf:             conf,
			}

			tc.checkExpectation(testAPI)

			r := incoming.NewRequest(proxy.conf, nil, "req-id").
				WithDestination("app1").
				WithPrevContext(apps.Context{
					ExpandedContext: apps.ExpandedContext{ActingUser: &model.User{Id: "userid"}},
				})
			r.Log = utils.NewTestLogger()
			_, resp := proxy.InvokeCall(r, tc.callRequest)

			require.Equal(t, tc.callResponse.RefreshBindings, resp.RefreshBindings)

			testAPI.AssertExpectations(t)
		})
	}
}
