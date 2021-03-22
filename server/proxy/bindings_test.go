// +build !e2e

package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_store"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_upstream"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/upstream"
)

const (
	testAppID1  = "test-appID1"
	testUserID  = "test-userID"
	testChannelID = "test-channelID"
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

func testBindings(appID apps.AppID, parent apps.Location, n string) []*apps.Binding {
	return []*apps.Binding{
		{
			AppID:    appID,
			Location: "/channel_header",
			Bindings: []*apps.Binding{
				{
					AppID:    appID,
					Location: apps.Location(fmt.Sprintf("id-%s", n)),
					Hint:     fmt.Sprintf("hint-%s", n),
				},
			},
			DependsOnChannel: false,
			DependsOnUser: false,
		},
		{
			AppID:    appID,
			Location: "/command",
			Bindings: []*apps.Binding{
				{
					AppID:    appID,
					Location: apps.Location(fmt.Sprintf("id-%s", n)),
					Hint:     fmt.Sprintf("hint-%s", n),
				},
			},
			DependsOnChannel: true,
			DependsOnUser: false,
		},
		{
			AppID:    appID,
			Location: "/post_menu",
			Bindings: []*apps.Binding{
				{
					AppID:    appID,
					Location: apps.Location(fmt.Sprintf("id-%s", n)),
					Hint:     fmt.Sprintf("hint-%s", n),
				},
			},
			DependsOnChannel: false,
			DependsOnUser: true,
		},
		{
			AppID:    appID,
			Location: "/in_post",
			Bindings: []*apps.Binding{
				{
					AppID:    appID,
					Location: apps.Location(fmt.Sprintf("id-%s", n)),
					Hint:     fmt.Sprintf("hint-%s", n),
				},
			},
			DependsOnChannel: true,
			DependsOnUser: true,
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
			require.EqualValues(t, tc.expected, out)
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

			proxy, _ := newTestProxyForBindings(testData, ctrl)

			cc := &apps.Context{}
			out, err := proxy.GetBindings("", cc)
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
							Location:    "ignored",
							Label:       "ignored",
							Icon:        "base command icon",
							Hint:        "base command hint",
							Description: "base command description",
							Bindings: []*apps.Binding{
								{
									Location:    "message",
									Label:       "message",
									Icon:        "message command icon",
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
									Icon:        "manage command icon",
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
							Location:    "ignored",
							Label:       "ignored",
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
					Location:    "app1",
					Label:       "app1",
					Icon:        "base command icon",
					Hint:        "base command hint",
					Description: "base command description",
					Bindings: []*apps.Binding{
						{
							AppID:       apps.AppID("app1"),
							Location:    "message",
							Label:       "message",
							Icon:        "message command icon",
							Hint:        "message command hint",
							Description: "message command description",
						}, {
							AppID:       apps.AppID("app1"),
							Location:    "message-modal",
							Label:       "message-modal",
							Icon:        "message-modal command icon",
							Hint:        "message-modal command hint",
							Description: "message-modal command description",
						}, {
							AppID:       apps.AppID("app1"),
							Location:    "manage",
							Label:       "manage",
							Icon:        "manage command icon",
							Hint:        "manage command hint",
							Description: "manage command description",
							Bindings: []*apps.Binding{
								{
									AppID:       apps.AppID("app1"),
									Location:    "subscribe",
									Label:       "subscribe",
									Icon:        "subscribe command icon",
									Hint:        "subscribe command hint",
									Description: "subscribe command description",
								}, {
									AppID:       apps.AppID("app1"),
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
				{
					AppID:       apps.AppID("app2"),
					Location:    "app2",
					Label:       "app2",
					Icon:        "app2 base command icon",
					Hint:        "app2 base command hint",
					Description: "app2 base command description",
					Bindings: []*apps.Binding{
						{
							AppID:       apps.AppID("app2"),
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
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	proxy, _ := newTestProxyForBindings(testData, ctrl)

	cc := &apps.Context{}
	out, err := proxy.GetBindings("", cc)
	require.NoError(t, err)
	require.EqualValues(t, expected, out)
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
							Location:    "",
							Label:       "",
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
					Location:    "app1",
					Label:       "app1",
					Icon:        "base command icon",
					Hint:        "base command hint",
					Description: "base command description",
					Bindings: []*apps.Binding{
						{
							AppID:    apps.AppID("app1"),
							Location: "sub1",
							Label:    "sub1",
							Icon:     "sub1 icon 1",
						},
					},
				},
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	proxy, _ := newTestProxyForBindings(testData, ctrl)

	cc := &apps.Context{}
	out, err := proxy.GetBindings("", cc)
	require.NoError(t, err)
	require.EqualValues(t, expected, out)
}

func newTestProxyForBindings(testData []bindingTestData, ctrl *gomock.Controller) (*Proxy, *plugintest.API) {
	testAPI := &plugintest.API{}
	testAPI.On("LogDebug", mock.Anything).Return(nil)
	mm := pluginapi.NewClient(testAPI)

	s := store.NewService(mm, config.NewTestConfigurator(&config.Config{}))
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
		reader := ioutil.NopCloser(bytes.NewReader(bb))

		up := mock_upstream.NewMockUpstream(ctrl)
		up.EXPECT().Roundtrip(gomock.Any(), gomock.Any()).Return(reader, nil)
		upstreams[test.app.Manifest.AppID] = up
		appStore.EXPECT().Get(test.app.AppID).Return(test.app, nil)
	}

	appStore.EXPECT().AsMap().Return(appList)

	conf := mock_config.NewMockService(ctrl)
	conf.EXPECT().GetMattermostConfig().Return(&model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL: model.NewString("test.mattermost.com"),
		},
	}).AnyTimes()

	p := &Proxy{
		mm:               mm,
		store:            s,
		builtinUpstreams: upstreams,
		conf:             conf,
	}

	return p, testAPI
}

func TestGetAllBindings(t *testing.T) {
	bindings := testBindings(testAppID1, apps.LocationCommand, "test")
	controller := gomock.NewController(t)
	defer controller.Finish()

	appID := apps.AppID(testAppID1)
	testData := []bindingTestData{
		{
			app: &apps.App{
				Manifest: apps.Manifest{
					AppID:       appID,
					AppType:     apps.AppTypeBuiltin,
					DisplayName: "App 1",
				},
			},
			bindings,
		},
	}
	p, api := newTestProxyForBindings(testData, controller)
	defer api.AssertExpectations(t)

	context1 := &apps.Context {
		AppID : appID,
		ActingUserID: testUserID,
		UserID: testUserID,
		ChannelID: testChannelID,
	}

	key1 := p.CacheBuildKey(KEY_ALL_USERS, KEY_ALL_CHANNELS)
	key2 := p.CacheBuildKey(testUserID, KEY_ALL_CHANNELS)
	key3 := p.CacheBuildKey(KEY_ALL_USERS, testChannelID)
	key4 := p.CacheBuildKey(testUserID, testChannelID)

	bindingBytes := []byte{}
	bindingsBytes := [][]byte{}
	bindingBytes, _ = json.Marshal(&bindings[appID][0])
	bindingsBytes = append(bindingsBytes, bindingBytes)
	api.On("AppsCacheGet", string(appID), key1).Return(bindingsBytes, nil)

	bindingsBytes = [][]byte{}
	bindingBytes, _ = json.Marshal(&bindings[appID][1])
	bindingsBytes = append(bindingsBytes, bindingBytes)
	api.On("AppsCacheGet", string(appID), key2).Return(bindingsBytes, nil)

	bindingsBytes = [][]byte{}
	bindingBytes, _ = json.Marshal(&bindings[appID][2])
	bindingsBytes = append(bindingsBytes, bindingBytes)
	api.On("AppsCacheGet", string(appID), key3).Return(bindingsBytes, nil)

	bindingsBytes = [][]byte{}
	bindingBytes, _ = json.Marshal(&bindings[appID][3])
	bindingsBytes = append(bindingsBytes, bindingBytes)
	api.On("AppsCacheGet", string(appID), key4).Return(bindingsBytes, nil)

	outBindings, err := p.CacheGetAll(context1, appID)
	require.Equal(t, outBindings, bindings[appID])

	assert.NoError(t, err)
}

func TestGetBindings(t *testing.T) {
	appBindings := testBindings(testAppID1, apps.LocationCommand, "test")
	controller := gomock.NewController(t)
	defer controller.Finish()

	appID := apps.AppID(testAppID1)
	testData := []bindingTestData{
		{
			app: &apps.App{
				Manifest: apps.Manifest{
					AppID:       appID,
					AppType:     apps.AppTypeBuiltin,
					DisplayName: "App 1",
				},
			},
			bindings,
		}
	}
	p, api := newTestProxyForBindings(appBindings, controller)
	defer api.AssertExpectations(t)

	appID := apps.AppID(testAppID1)
	context1 := &apps.Context {
		AppID : appID,
		ActingUserID: testUserID,
		UserID: testUserID,
		ChannelID: testChannelID,
	}

	key2 := p.CacheBuildKey(testUserID, KEY_ALL_CHANNELS)

	bindingsBytes := [][]byte{}
	bindingBytes, _ := json.Marshal(&appBindings[appID][1])
	bindingsBytes = append(bindingsBytes, bindingBytes)
	api.On("AppsCacheGet", string(appID), key2).Return(bindingsBytes, nil)

	outBindings, err := p.CacheGet(context1, appID, key2)
	bindings := make([]*apps.Binding,0)
	bindings = append(bindings, appBindings[appID][1])
	require.Equal(t, outBindings, bindings)

	assert.NoError(t, err)
}

func TestDeleteBindingsForApp(t *testing.T) {
	bindings := testBindings(testAppID1, apps.LocationCommand, "test")
	controller := gomock.NewController(t)
	defer controller.Finish()

	appID := apps.AppID(testAppID1)
	testData := []bindingTestData{
		{
			app: &apps.App{
				Manifest: apps.Manifest{
					AppID:       appID,
					AppType:     apps.AppTypeBuiltin,
					DisplayName: "App 1",
				},
			},
			bindings,
		}
	}
	p, api := newTestProxyForBindings(testData, controller)
	defer api.AssertExpectations(t)

	key1 := p.CacheBuildKey(KEY_ALL_USERS, KEY_ALL_CHANNELS)

	api.On("AppsCacheDelete", string(appID), key1).Return(nil)
	err := p.CacheDelete(appID, key1)

	assert.NoError(t, err)
}

func TestDeleteAllBindings(t *testing.T) {
	appBindings := testBindings(testAppID1, apps.LocationCommand, "test")
	controller := gomock.NewController(t)
	defer controller.Finish()

	appID := apps.AppID(testAppID1)
	testData := []bindingTestData{
		{
			app: &apps.App{
				Manifest: apps.Manifest{
					AppID:       appID,
					AppType:     apps.AppTypeBuiltin,
					DisplayName: "App 1",
				},
			},
			bindings,
		}
	}
	p, api := newTestProxyForBindings(testData, controller)
	defer api.AssertExpectations(t)

	api.On("AppsCacheDeleteAll", string(appID)).Return(nil)
	err := p.CacheDeleteAll(appID)

	assert.NoError(t, err)
}

func TestDeleteAllBindingsForAllApps(t *testing.T) {
	bindings := testBindings(testAppID1, apps.LocationCommand, "test")
	controller := gomock.NewController(t)
	defer controller.Finish()

	appID := apps.AppID(testAppID1)
	testData := []bindingTestData{
		{
			app: &apps.App{
				Manifest: apps.Manifest{
					AppID:       appID,
					AppType:     apps.AppTypeBuiltin,
					DisplayName: "App 1",
				},
			},
			bindings,
		},
	}
	p, api := newTestProxyForBindings(testData, controller)
	defer api.AssertExpectations(t)

	appID := apps.AppID(testAppID1)

	s := mock_api.NewMockStore(controller)
	appStore := mock_api.NewMockAppStore(controller)
	s.EXPECT().App().Return(appStore)
	p.store = s

	appList := []*apps.App{}
	for k, _ := range appBindings {
		tapp := &apps.App{
			Manifest: &apps.Manifest{
				AppID:              apps.AppID(k),
				Type:               apps.AppTypeBuiltin,
			},
		}
		appList = append(appList, tapp)
	}
	appStore.EXPECT().GetAll().Return(appList)

	api.On("AppsCacheDeleteAll", string(appID)).Return(nil)
	errors := p.CacheDeleteAllApps()

	assert.Equal(t, len(errors), 0)
}

func TestSetBindings(t *testing.T) {
	bindings := testBindings(testAppID1, apps.LocationCommand, "test")
	controller := gomock.NewController(t)
	defer controller.Finish()

	appID := apps.AppID(testAppID1)
	testData := []bindingTestData{
		{
			app: &apps.App{
				Manifest: apps.Manifest{
					AppID:       appID,
					AppType:     apps.AppTypeBuiltin,
					DisplayName: "App 1",
				},
			},
			bindings,
		},
	}
	p, api := newTestProxyForBindings(testData, controller)
	defer api.AssertExpectations(t)

	appID := apps.AppID(testAppID1)
	context1 := &apps.Context {
		AppID : appID,
		ActingUserID: testUserID,
		UserID: testUserID,
		ChannelID: testChannelID,
	}

	var key string
	bindingsMap := make(map[string][][]byte, len(bindings[appID]))

	for _, binding := range bindings[appID] {
		if !binding.DependsOnUser && !binding.DependsOnChannel {
			key = p.CacheBuildKey(KEY_ALL_USERS, KEY_ALL_CHANNELS)
		} else if binding.DependsOnUser && !binding.DependsOnChannel {
			key = p.CacheBuildKey(testUserID, KEY_ALL_CHANNELS)
		} else if !binding.DependsOnUser && binding.DependsOnChannel {
			key = p.CacheBuildKey(KEY_ALL_USERS, testChannelID)
		} else {
			key = p.CacheBuildKey(testUserID, testChannelID)
		}
		valueBytes, _ := json.Marshal(&binding)

		bindingsMap[key] = make([][]byte, 0)
		bindingsForKey := bindingsMap[key]
		bindingsForKey = append(bindingsForKey, valueBytes)
		bindingsMap[key] = bindingsForKey

		api.On("AppsCacheSet", string(appID), bindingsMap).Return(nil)
	}

	err := p.CacheSet(context1, appID, bindings[appID])

	assert.NoError(t, err)
}
