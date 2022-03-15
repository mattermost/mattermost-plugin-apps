package proxy

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func TestAppMetadataForClient(t *testing.T) {
	testApps := []apps.App{
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

	ctrl := gomock.NewController(t)

	proxy := newTestProxy(t, testApps, ctrl)
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

	r := incoming.NewRequest(proxy.conf.MattermostAPI(), proxy.conf, utils.NewTestLogger(), nil)
	r.Log = utils.NewTestLogger()
	resp := proxy.Call(r, creq)
	require.Equal(t, resp.AppMetadata, AppMetadataForClient{
		BotUserID:   "botid",
		BotUsername: "botusername",
	})
}
