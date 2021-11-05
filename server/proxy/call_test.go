package proxy

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy/request"
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
	defer ctrl.Finish()

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

	c := request.NewContext(nil, p.conf, nil)
	c.Log = utils.NewTestLogger()
	resp := p.Call(c, creq)
	require.Equal(t, resp.AppMetadata, AppMetadataForClient{
		BotUserID:   "botid",
		BotUsername: "botusername",
	})
}
