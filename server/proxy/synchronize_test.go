package proxy

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/golang/mock/gomock"
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/cluster"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_upstream"
	mock_detector "github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_upstream_detector"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

func TestSynchronize(t *testing.T) {
	type TC struct {
		name           string
		inputManifests map[apps.AppID]*apps.Manifest
		inputApps      map[apps.AppID]*apps.App
		needUpstream   bool

		expectedManifests map[apps.AppID]*apps.Manifest
		expectedApps      map[apps.AppID]*apps.App
	}

	for _, tc := range []TC{
		{
			name:           "no installed, no listed",
			inputApps:      map[apps.AppID]*apps.App{},
			inputManifests: map[apps.AppID]*apps.Manifest{},
			needUpstream:   false,

			expectedApps:      map[apps.AppID]*apps.App{},
			expectedManifests: map[apps.AppID]*apps.Manifest{},
		},
		{
			name:      "no installed, one listed",
			inputApps: map[apps.AppID]*apps.App{},
			inputManifests: map[apps.AppID]*apps.Manifest{
				"id": {
					AppID:   "id",
					AppType: "http",
				},
			},
			needUpstream: false,

			expectedApps: map[apps.AppID]*apps.App{},
			expectedManifests: map[apps.AppID]*apps.Manifest{
				"id": {
					AppID:   "id",
					AppType: "http",
				},
			},
		},
		{
			name: "one listed, one unlisted",
			inputApps: map[apps.AppID]*apps.App{
				"id1": {
					Manifest: apps.Manifest{
						AppID:            "id1",
						AppType:          "http",
						Version:          "1",
						OnVersionChanged: &apps.Call{},
					},
					Disabled: false,
				},
			},
			inputManifests: map[apps.AppID]*apps.Manifest{
				"id2": {
					AppID:   "id2",
					AppType: "http",
				},
			},
			needUpstream: false,

			expectedApps: map[apps.AppID]*apps.App{
				"id1": {
					Manifest: apps.Manifest{
						AppID:            "id1",
						AppType:          "http",
						Version:          "1",
						OnVersionChanged: &apps.Call{},
					},
					Disabled: false,
				},
			},
			expectedManifests: map[apps.AppID]*apps.Manifest{
				"id2": {
					AppID:   "id2",
					AppType: "http",
				},
			},
		},
		{
			name: "update version",
			inputApps: map[apps.AppID]*apps.App{
				"id1": {
					Manifest: apps.Manifest{
						AppID:            "id1",
						AppType:          "http",
						Version:          "1",
						OnVersionChanged: &apps.Call{},
					},
					Disabled: false,
				},
			},
			inputManifests: map[apps.AppID]*apps.Manifest{
				"id1": {
					AppID:            "id1",
					AppType:          "http",
					Version:          "2",
					OnVersionChanged: &apps.Call{},
				},
			},
			needUpstream: true,

			expectedApps: map[apps.AppID]*apps.App{
				"id1": {
					Manifest: apps.Manifest{
						AppID:            "id1",
						AppType:          "http",
						Version:          "2",
						OnVersionChanged: &apps.Call{},
					},
					Disabled: false,
				},
			},
			expectedManifests: map[apps.AppID]*apps.Manifest{
				"id1": {
					AppID:            "id1",
					AppType:          "http",
					Version:          "2",
					OnVersionChanged: &apps.Call{},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := createProxy(tc.inputApps, tc.inputManifests)

			if tc.needUpstream {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				cr := &apps.CallResponse{
					Type: apps.CallResponseTypeOK,
					Data: "some data",
				}
				bb, _ := json.Marshal(cr)
				reader := ioutil.NopCloser(bytes.NewReader(bb))
				up := mock_upstream.NewMockUpstream(ctrl)
				up.EXPECT().Roundtrip(gomock.Any(), gomock.Any()).Return(reader, nil)

				upDetector := mock_detector.NewMockDetector(ctrl)
				upDetector.EXPECT().UpstreamForApp(gomock.Any()).Return(up, nil)

				p.upstreamDetector = upDetector
			}

			err := p.SynchronizeInstalledApps()
			require.NoError(t, err)
			installedApps := p.store.App.AsMap()
			require.Equal(t, tc.expectedApps, installedApps)
			listedApps := p.store.Manifest.AsMap()
			require.Equal(t, tc.expectedManifests, listedApps)
			for _, app := range installedApps {
				if _, isListed := listedApps[app.AppID]; !isListed {
					require.False(t, p.AppIsEnabled(app))
				}
			}
		})
	}
}

func createProxy(a map[apps.AppID]*apps.App, m map[apps.AppID]*apps.Manifest) *Proxy {
	testAPI := &plugintest.API{}
	testAPI.On("LogDebug", mock.Anything).Return(nil)
	testAPI.On("KVSetWithOptions", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
	testAPI.On("KVGet", mock.Anything).Return(nil, nil)

	mm := pluginapi.NewClient(testAPI)

	s := store.NewService(mm, nil)
	appService := store.NewAppStoreMock(a)
	manifestService := store.NewManifestStoreMock(m)
	s.App = appService
	s.Manifest = manifestService

	mutex, _ := cluster.NewMutex(testAPI, config.KeyClusterMutex)

	conf := config.NewTestConfigurator(config.Config{}).WithMattermostConfig(model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL: model.NewString("test.mattermost.com"),
		},
	})

	return &Proxy{
		mm:            mm,
		store:         s,
		callOnceMutex: mutex,
		conf:          conf,
	}
}
