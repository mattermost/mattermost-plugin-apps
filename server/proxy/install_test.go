// +build !e2e

package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_clients"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_store"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_upstream"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/upstream"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func TestInstallApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testAPI := &plugintest.API{}
	testAPI.On("LogDebug", mock.Anything).Return(nil)
	mm := pluginapi.NewClient(testAPI)

	siteURL := "http://test.mattermost.com"
	conf := config.NewTestConfigurator(config.Config{
		MattermostSiteURL: siteURL,
	}).WithMattermostConfig(model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL: model.NewString(siteURL),
		},
	})

	s := store.NewService(mm, conf)
	appStore := mock_store.NewMockAppStore(ctrl)
	s.App = appStore

	appID := apps.AppID("app1")
	manifest := apps.Manifest{
		AppID:       apps.AppID("app1"),
		AppType:     apps.AppTypeBuiltin,
		DisplayName: "App 1",
		RequestedPermissions: apps.Permissions{
			apps.PermissionActAsAdmin,
		},
	}

	app := &apps.App{
		Manifest: manifest,
		GrantedLocations: apps.Locations{
			apps.LocationCommand,
		},
	}

	appStore.EXPECT().Get(appID).Times(2).Return(app, nil)
	appStore.EXPECT().Save(app).Return(nil)

	upstreams := map[apps.AppID]upstream.Upstream{}
	up := mock_upstream.NewMockUpstream(ctrl)
	cr := &apps.CallResponse{
		Markdown: "Install done",
	}
	bb, _ := json.Marshal(cr)
	reader := io.NopCloser(bytes.NewReader(bb))
	up.EXPECT().Roundtrip(gomock.Any(), gomock.Any()).Return(reader, nil)
	upstreams[appID] = up

	clientService := mock_clients.NewMockClientService(ctrl)
	client := mock_clients.NewMockClient4(ctrl)
	clientService.EXPECT().NewClient4(siteURL).Return(client)

	manifestStore := mock_store.NewMockManifestStore(ctrl)
	manifestStore.EXPECT().Get(appID).Return(&manifest, nil)
	s.Manifest = manifestStore

	p := &Proxy{
		mm:               mm,
		store:            s,
		conf:             conf,
		clients:          clientService,
		builtinUpstreams: upstreams,
	}

	session := &model.Session{
		Id:     "sessionid",
		UserId: "actinguserid",
		Token:  "thetoken",
	}
	testAPI.On("GetSession", "sessionid").Return(session, nil)
	testAPI.On("HasPermissionTo", "actinguserid", model.PERMISSION_MANAGE_SYSTEM).Return(true)

	client.EXPECT().SetToken("thetoken")
	client.EXPECT().GetUserByUsername(string(appID), "").Return(nil, nil)
	client.EXPECT().CreateBot(gomock.Any()).Return(&model.Bot{UserId: "botuserid", Username: "botusername"}, &model.Response{StatusCode: http.StatusCreated})

	testAPI.On("GetDirectChannel", "botuserid", "actinguserid").Return(&model.Channel{Id: "channelid"}, nil)
	testAPI.On("CreatePost", &model.Post{
		UserId:    "botuserid",
		ChannelId: "channelid",
		Message:   "Using bot account @botusername (`botuserid`).",
	}).Return(&model.Post{}, nil)

	testAPI.On("PublishWebSocketEvent", "refresh_bindings", map[string]interface{}{}, mock.AnythingOfType("*model.WebsocketBroadcast"))

	cc := &apps.Context{
		AppID:        appID,
		ActingUserID: "actinguserid",
	}
	trusted := true
	app, out, err := p.InstallApp("sessionid", "actinguserid", cc, trusted, "secret")
	require.NoError(t, err)
	require.NotNil(t, app)
	require.Equal(t, out, md.MD("Install done"))
}

func TestEnsureBot(t *testing.T) {
	testData := []struct {
		name        string
		manifest    apps.Manifest
		createToken bool
	}{
		{
			name:        "act_as_bot requested",
			createToken: true,
			manifest: apps.Manifest{
				AppID:       apps.AppID("app1"),
				AppType:     apps.AppTypeBuiltin,
				DisplayName: "App 1",
				RequestedPermissions: apps.Permissions{
					apps.PermissionActAsBot,
				},
			},
		},
		{
			name:        "act_as_bot not requested",
			createToken: false,
			manifest: apps.Manifest{
				AppID:                apps.AppID("app1"),
				AppType:              apps.AppTypeBuiltin,
				DisplayName:          "App 1",
				RequestedPermissions: apps.Permissions{},
			},
		},
	}

	for _, tc := range testData {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			appID := tc.manifest.AppID
			manifest := tc.manifest

			testAPI := &plugintest.API{}
			testAPI.On("LogDebug", mock.Anything).Return(nil)
			mm := pluginapi.NewClient(testAPI)

			siteURL := "http://test.mattermost.com"
			conf := config.NewTestConfigurator(config.Config{
				MattermostSiteURL: siteURL,
			}).WithMattermostConfig(model.Config{
				ServiceSettings: model.ServiceSettings{
					SiteURL: model.NewString(siteURL),
				},
			})

			s := store.NewService(mm, conf)
			appStore := mock_store.NewMockAppStore(ctrl)
			s.App = appStore

			client := mock_clients.NewMockClient4(ctrl)

			p := &Proxy{
				mm:    mm,
				store: s,
				conf:  conf,
			}
			testAPI.On("HasPermissionTo", "actinguserid", model.PERMISSION_MANAGE_SYSTEM).Return(true)

			client.EXPECT().GetUserByUsername(string(appID), "").Return(nil, nil)
			client.EXPECT().CreateBot(gomock.Any()).Return(&model.Bot{UserId: "botuserid"}, &model.Response{StatusCode: http.StatusCreated})

			testAPI.On("GetDirectChannel", "botuserid", "actinguserid").Return(&model.Channel{}, nil)
			testAPI.On("CreatePost", mock.AnythingOfType("*model.Post")).Return(&model.Post{}, nil)

			if tc.createToken {
				client.EXPECT().GetUser("botuserid", "").Return(&model.User{Roles: "system_user"}, &model.Response{StatusCode: http.StatusOK})
				client.EXPECT().UpdateUserRoles("botuserid", "system_user system_post_all").Return(true, &model.Response{StatusCode: http.StatusOK})
				client.EXPECT().CreateUserAccessToken("botuserid", "Mattermost App Token").Return(&model.UserAccessToken{}, &model.Response{StatusCode: http.StatusOK})
			}

			bot, token, err := p.ensureBot(&manifest, "actinguserid", client)
			require.NoError(t, err)
			require.NotNil(t, bot)
			require.Equal(t, token, token)
		})
	}
}
