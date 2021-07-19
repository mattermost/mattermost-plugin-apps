package proxy

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/golang/mock/gomock"
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

func TestAppMetadataForClient(t *testing.T) {
	testApps := []*apps.App{
		{
			BotUserID:   "botid",
			BotUsername: "botusername",
			Manifest: apps.Manifest{
				AppID:       apps.AppID("app1"),
				AppType:     apps.AppTypeBuiltin,
				DisplayName: "App 1",
			},
		},
	}

	ctrl := gomock.NewController(t)
	p := newTestProxy(testApps, ctrl)
	c := &apps.CallRequest{
		Context: &apps.Context{
			UserAgentContext: apps.UserAgentContext{
				AppID: "app1",
			},
		},
		Call: apps.Call{
			Path: "/",
		},
	}

	resp := p.Call("session_id", "acting_user_id", c)
	require.Equal(t, resp.AppMetadata, &apps.AppMetadataForClient{
		BotUserID:   "botid",
		BotUsername: "botusername",
	})
}

func newTestProxy(testApps []*apps.App, ctrl *gomock.Controller) *Proxy {
	testAPI := &plugintest.API{}
	testAPI.On("LogDebug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mm := pluginapi.NewClient(testAPI)

	conf := config.NewTestConfigurator(config.Config{}).WithMattermostConfig(model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL: model.NewString("test.mattermost.com"),
		},
	})

	s, _ := store.MakeService(mm, conf, nil)
	appStore := mock_store.NewMockAppStore(ctrl)
	s.App = appStore

	upstreams := map[apps.AppID]upstream.Upstream{}
	for _, app := range testApps {
		cr := &apps.CallResponse{
			Type: apps.CallResponseTypeOK,
		}
		b, _ := json.Marshal(cr)
		reader := ioutil.NopCloser(bytes.NewReader(b))

		up := mock_upstream.NewMockUpstream(ctrl)
		up.EXPECT().Roundtrip(gomock.Any(), gomock.Any(), gomock.Any()).Return(reader, nil)
		upstreams[app.Manifest.AppID] = up
		appStore.EXPECT().Get(app.AppID).Return(app, nil)
	}

	p := &Proxy{
		mm:               mm,
		store:            s,
		builtinUpstreams: upstreams,
		conf:             conf,
	}

	return p
}
