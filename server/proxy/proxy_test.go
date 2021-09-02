package proxy

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_store"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_upstream"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
)

func TestAppMetadataForClient(t *testing.T) {
	testApps := []apps.App{
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
	p := newTestProxy(t, testApps, ctrl)
	creq := apps.CallRequest{
		Context: apps.Context{
			UserAgentContext: apps.UserAgentContext{
				AppID: "app1",
			},
		},
		Call: apps.Call{
			Path: "/",
		},
	}

	resp := p.Call(Incoming{}, creq)
	require.Equal(t, resp.AppMetadata, &apps.AppMetadataForClient{
		BotUserID:   "botid",
		BotUsername: "botusername",
	})
}

func newTestProxy(tb testing.TB, testApps []apps.App, ctrl *gomock.Controller) *Proxy {
	conf := config.NewTestConfigService(nil).WithMattermostConfig(model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL: model.NewString("test.mattermost.com"),
		},
	})

	s, err := store.MakeService(conf, nil)
	require.NoError(tb, err)
	appStore := mock_store.NewMockAppStore(ctrl)
	s.App = appStore

	upstreams := map[apps.AppID]upstream.Upstream{}
	for i := range testApps {
		app := testApps[i]

		up := mock_upstream.NewMockUpstream(ctrl)

		// set up an empty OK call response
		b, _ := json.Marshal(apps.CallResponse{
			Type: apps.CallResponseTypeOK,
		})
		reader := ioutil.NopCloser(bytes.NewReader(b))
		up.EXPECT().Roundtrip(gomock.Any(), gomock.Any(), gomock.Any()).Return(reader, nil)

		upstreams[app.Manifest.AppID] = up
		appStore.EXPECT().Get(app.AppID).Return(&app, nil)
	}

	p := &Proxy{
		store:            s,
		builtinUpstreams: upstreams,
		conf:             conf,
	}

	return p
}
