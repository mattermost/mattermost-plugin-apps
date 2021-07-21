// +build !e2e

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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_store"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_upstream"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
)

type bindingTestData struct {
	app      *apps.App
	bindings []*apps.Binding
}

func testBinding(appID apps.AppID, parent apps.Location, n string) []*apps.Binding {
	return []*apps.Binding{
		{
			AppID:    appID,
			Location: parent,
			Bindings: []*apps.Binding{
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
		bb1, bb2, expected []*apps.Binding
	}

	for _, tc := range []TC{
		{
			name: "happy simplest",
			bb1: []*apps.Binding{
				{
					Location: "1",
				},
			},
			bb2: []*apps.Binding{
				{
					Location: "2",
				},
			},
			expected: []*apps.Binding{
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
			expected: []*apps.Binding{
				{
					AppID:    "app1",
					Location: "/command",
					Bindings: []*apps.Binding{
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
			bb1: []*apps.Binding{
				{
					Location: "/post_menu",
					Bindings: []*apps.Binding{
						{
							AppID:       "zendesk",
							Label:       "Create zendesk ticket",
							Description: "Create ticket in zendesk",
							Call: &apps.Call{
								Path: "http://localhost:4000/create",
							},
						},
					},
				},
			},
			bb2: []*apps.Binding{
				{
					Location: "/post_menu",
					Bindings: []*apps.Binding{
						{
							AppID:       "hello",
							Label:       "Create hello ticket",
							Description: "Create ticket in hello",
							Call: &apps.Call{
								Path: "http://localhost:4000/hello",
							},
						},
					},
				},
			},
			expected: []*apps.Binding{
				{
					Location: "/post_menu",
					Bindings: []*apps.Binding{
						{
							AppID:       "zendesk",
							Label:       "Create zendesk ticket",
							Description: "Create ticket in zendesk",
							Call: &apps.Call{
								Path: "http://localhost:4000/create",
							},
						},
						{
							AppID:       "hello",
							Label:       "Create hello ticket",
							Description: "Create ticket in hello",
							Call: &apps.Call{
								Path: "http://localhost:4000/hello",
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := mergeBindings(tc.bb1, tc.bb2)
			EqualBindings(t, tc.expected, out)
		})
	}
}

func TestGetBindingsGrantedLocations(t *testing.T) {
	type TC struct {
		name        string
		locations   apps.Locations
		numBindings int
	}

	for _, tc := range []TC{
		{
			name: "3 locations granted",
			locations: apps.Locations{
				apps.LocationChannelHeader,
				apps.LocationPostMenu,
				apps.LocationCommand,
			},
			numBindings: 3,
		},
		{
			name: "command location granted",
			locations: apps.Locations{
				apps.LocationCommand,
			},
			numBindings: 1,
		},
		{
			name: "channel header location granted",
			locations: apps.Locations{
				apps.LocationChannelHeader,
			},
			numBindings: 1,
		},
		{
			name: "post dropdown location granted",
			locations: apps.Locations{
				apps.LocationPostMenu,
			},
			numBindings: 1,
		},
		{
			name:        "no granted locations",
			locations:   apps.Locations{},
			numBindings: 0,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bindings := []*apps.Binding{
				{
					Location: apps.LocationChannelHeader,
					Bindings: []*apps.Binding{
						{
							Location: "send",
							Label:    "Send",
						},
					},
				}, {
					Location: apps.LocationPostMenu,
					Bindings: []*apps.Binding{
						{
							Location: "send-me",
							Label:    "Send me",
						},
					},
				}, {
					Location: apps.LocationCommand,
					Bindings: []*apps.Binding{
						{
							Location: "ignored",
							Label:    "ignored",
						},
					},
				},
			}

			app1 := &apps.App{
				Manifest: apps.Manifest{
					AppID:              apps.AppID("app1"),
					AppType:            apps.AppTypeBuiltin,
					DisplayName:        "App 1",
					RequestedLocations: tc.locations,
				},
				GrantedLocations: tc.locations,
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			testData := []bindingTestData{{
				app:      app1,
				bindings: bindings,
			}}

			proxy := newTestProxyForBindings(testData, ctrl)

			cc := &apps.Context{}
			out, err := proxy.GetBindings("", "", cc)
			require.NoError(t, err)
			require.Len(t, out, tc.numBindings)
		})
	}
}

func TestGetBindingsCommands(t *testing.T) {
	testData := []bindingTestData{
		{
			app: &apps.App{
				Manifest: apps.Manifest{
					AppID:       apps.AppID("app1"),
					AppType:     apps.AppTypeBuiltin,
					DisplayName: "App 1",
				},
				GrantedLocations: apps.Locations{
					apps.LocationChannelHeader,
					apps.LocationPostMenu,
					apps.LocationCommand,
				},
			},
			bindings: []*apps.Binding{
				{
					Location: apps.LocationCommand,
					Bindings: []*apps.Binding{
						{
							Location:    "base command location",
							Label:       "base command label",
							Icon:        "base command icon",
							Hint:        "base command hint",
							Description: "base command description",
							Bindings: []*apps.Binding{
								{
									Location:    "message",
									Label:       "message",
									Icon:        "https://example.com/image.png",
									Hint:        "message command hint",
									Description: "message command description",
								}, {
									Location:    "message-modal",
									Label:       "message-modal",
									Icon:        "message-modal command icon",
									Hint:        "message-modal command hint",
									Description: "message-modal command description",
								}, {
									Location:    "manage",
									Label:       "manage",
									Icon:        "../some/invalid/path",
									Hint:        "manage command hint",
									Description: "manage command description",
									Bindings: []*apps.Binding{
										{
											Location:    "subscribe",
											Label:       "subscribe",
											Icon:        "subscribe command icon",
											Hint:        "subscribe command hint",
											Description: "subscribe command description",
										}, {
											Location:    "unsubscribe",
											Label:       "unsubscribe",
											Icon:        "unsubscribe command icon",
											Hint:        "unsubscribe command hint",
											Description: "unsubscribe command description",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			app: &apps.App{
				Manifest: apps.Manifest{
					AppID:       apps.AppID("app2"),
					AppType:     apps.AppTypeBuiltin,
					DisplayName: "App 2",
				},
				GrantedLocations: apps.Locations{
					apps.LocationChannelHeader,
					apps.LocationPostMenu,
					apps.LocationCommand,
				},
			},
			bindings: []*apps.Binding{
				{
					Location: apps.LocationCommand,
					Bindings: []*apps.Binding{
						{
							Location:    "app2 base command location",
							Label:       "app2 base command label",
							Icon:        "app2 base command icon",
							Hint:        "app2 base command hint",
							Description: "app2 base command description",
							Bindings: []*apps.Binding{
								{
									Location:    "connect",
									Label:       "connect",
									Icon:        "connect command icon",
									Hint:        "connect command hint",
									Description: "connect command description",
								},
							},
						},
					},
				},
			},
		},
	}

	expected := []*apps.Binding{
		{
			Location: apps.LocationCommand,
			Bindings: []*apps.Binding{
				{
					AppID:       apps.AppID("app1"),
					Location:    "base command location",
					Label:       "base command label",
					Icon:        "https://test.mattermost.com/plugins/com.mattermost.apps/apps/app1/static/base command icon",
					Hint:        "base command hint",
					Description: "base command description",
					Bindings: []*apps.Binding{
						{
							AppID:       apps.AppID("app1"),
							Location:    "message",
							Label:       "message",
							Icon:        "https://example.com/image.png",
							Hint:        "message command hint",
							Description: "message command description",
						}, {
							AppID:       apps.AppID("app1"),
							Location:    "message-modal",
							Label:       "message-modal",
							Icon:        "https://test.mattermost.com/plugins/com.mattermost.apps/apps/app1/static/message-modal command icon",
							Hint:        "message-modal command hint",
							Description: "message-modal command description",
						}, {
							AppID:       apps.AppID("app1"),
							Location:    "manage",
							Label:       "manage",
							Icon:        "",
							Hint:        "manage command hint",
							Description: "manage command description",
							Bindings: []*apps.Binding{
								{
									AppID:       apps.AppID("app1"),
									Location:    "subscribe",
									Label:       "subscribe",
									Icon:        "https://test.mattermost.com/plugins/com.mattermost.apps/apps/app1/static/subscribe command icon",
									Hint:        "subscribe command hint",
									Description: "subscribe command description",
								}, {
									AppID:       apps.AppID("app1"),
									Location:    "unsubscribe",
									Label:       "unsubscribe",
									Icon:        "https://test.mattermost.com/plugins/com.mattermost.apps/apps/app1/static/unsubscribe command icon",
									Hint:        "unsubscribe command hint",
									Description: "unsubscribe command description",
								},
							},
						},
					},
				},
				{
					AppID:       apps.AppID("app2"),
					Location:    "app2 base command location",
					Label:       "app2 base command label",
					Icon:        "https://test.mattermost.com/plugins/com.mattermost.apps/apps/app2/static/app2 base command icon",
					Hint:        "app2 base command hint",
					Description: "app2 base command description",
					Bindings: []*apps.Binding{
						{
							AppID:       apps.AppID("app2"),
							Location:    "connect",
							Label:       "connect",
							Icon:        "https://test.mattermost.com/plugins/com.mattermost.apps/apps/app2/static/connect command icon",
							Hint:        "connect command hint",
							Description: "connect command description",
						},
					},
				},
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	proxy := newTestProxyForBindings(testData, ctrl)

	cc := &apps.Context{}
	out, err := proxy.GetBindings("", "", cc)
	require.NoError(t, err)
	EqualBindings(t, expected, out)
}

func TestDuplicateCommand(t *testing.T) {
	testData := []bindingTestData{
		{
			app: &apps.App{
				Manifest: apps.Manifest{
					AppID:       apps.AppID("app1"),
					AppType:     apps.AppTypeBuiltin,
					DisplayName: "App 1",
				},
				GrantedLocations: apps.Locations{
					apps.LocationCommand,
				},
			},
			bindings: []*apps.Binding{
				{
					Location: apps.LocationCommand,
					Bindings: []*apps.Binding{
						{
							Location:    "base command location",
							Label:       "base command label",
							Icon:        "base command icon",
							Hint:        "base command hint",
							Description: "base command description",
							Bindings: []*apps.Binding{
								{
									Location: "sub1",
									Label:    "sub1",
									Icon:     "sub1 icon 1",
								},
								{
									Location: "sub1",
									Label:    "sub1",
									Icon:     "sub1 icon 2",
								},
								{
									Location: "",
									Label:    "",
									Icon:     "",
								},
							},
						},
					},
				},
				{
					Location: apps.LocationCommand,
					Bindings: []*apps.Binding{
						{
							Location:    "",
							Label:       "",
							Icon:        "base2 command icon",
							Hint:        "base2 command hint",
							Description: "base2 command description",
						},
					},
				},
			},
		},
	}

	expected := []*apps.Binding{
		{
			Location: apps.LocationCommand,
			Bindings: []*apps.Binding{
				{
					AppID:       apps.AppID("app1"),
					Location:    "base command location",
					Label:       "base command label",
					Icon:        "https://test.mattermost.com/plugins/com.mattermost.apps/apps/app1/static/base command icon",
					Hint:        "base command hint",
					Description: "base command description",
					Bindings: []*apps.Binding{
						{
							AppID:    apps.AppID("app1"),
							Location: "sub1",
							Label:    "sub1",
							Icon:     "https://test.mattermost.com/plugins/com.mattermost.apps/apps/app1/static/sub1 icon 1",
						},
					},
				},
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	proxy := newTestProxyForBindings(testData, ctrl)

	cc := &apps.Context{}
	out, err := proxy.GetBindings("", "", cc)
	require.NoError(t, err)
	EqualBindings(t, expected, out)
}

func newTestProxyForBindings(testData []bindingTestData, ctrl *gomock.Controller) *Proxy {
	testAPI := &plugintest.API{}
	testAPI.On("LogDebug", mock.AnythingOfType("string")).Return(nil)
	testAPI.On("LogDebug", mock.AnythingOfType("string"),
		mock.Anything, mock.Anything,
		mock.Anything, mock.Anything,
		mock.Anything, mock.Anything,
	).Return(nil)
	testDriver := &plugintest.Driver{}
	mm := pluginapi.NewClient(testAPI, testDriver)

	conf := config.Config{
		PluginURL: "https://test.mattermost.com/plugins/com.mattermost.apps",
	}
	confService := config.NewTestConfigurator(conf).WithMattermostConfig(model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL: model.NewString("https://test.mattermost.com"),
		},
	})

	s := store.NewService(mm, confService, nil, "")
	appStore := mock_store.NewMockAppStore(ctrl)
	s.App = appStore

	appList := map[apps.AppID]*apps.App{}
	upstreams := map[apps.AppID]upstream.Upstream{}

	for _, test := range testData {
		appList[test.app.AppID] = test.app

		cr := &apps.CallResponse{
			Type: apps.CallResponseTypeOK,
			Data: test.bindings,
		}
		bb, _ := json.Marshal(cr)
		reader := io.NopCloser(bytes.NewReader(bb))

		up := mock_upstream.NewMockUpstream(ctrl)
		up.EXPECT().Roundtrip(gomock.Any(), gomock.Any()).Return(reader, nil)
		upstreams[test.app.Manifest.AppID] = up
		appStore.EXPECT().Get(test.app.AppID).Return(test.app, nil)
	}

	appStore.EXPECT().AsMap().Return(appList)

	p := &Proxy{
		mm:               mm,
		store:            s,
		builtinUpstreams: upstreams,
		conf:             confService,
	}

	return p
}

// EqualBindings asserts that two slices of bindings are equal ignoring the order of the elements.
// If there are duplicate elements, the number of appearances of each of them in both lists should match.
//
// EqualBindings calls t.Fail if the elements not match.
func EqualBindings(t *testing.T, expected, actual []*apps.Binding) {
	opt := cmpopts.SortSlices(func(a *apps.Binding, b *apps.Binding) bool {
		return a.AppID < b.AppID
	})

	if diff := cmp.Diff(expected, actual, opt); diff != "" {
		t.Errorf("Bindings mismatch (-expected +actual):\n%s", diff)
	}
}
