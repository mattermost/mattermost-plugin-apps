package restapi

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_proxy"
)

func TestHandleGetBindingsInvalidContext(t *testing.T) {
	ctrl := gomock.NewController(t)

	proxy := mock_proxy.NewMockService(ctrl)

	api := &restapi{
		proxy: proxy,
	}

	cc := &apps.Context{
		ContextFromUserAgent: apps.ContextFromUserAgent{
			TeamID: "some_team_id",
		},
	}

	proxy.EXPECT().CleanUserCallContext("some_user_id", cc).Return(nil, errors.New("user is not a member of the specified team"))

	res, err := api.handleGetBindings("some_session_id", "some_user_id", cc)
	require.Error(t, err)
	require.Nil(t, res)
}

func TestHandleGetBindingsValidContext(t *testing.T) {
	ctrl := gomock.NewController(t)

	proxy := mock_proxy.NewMockService(ctrl)
	conf := mock_config.NewMockService(ctrl)

	api := &restapi{
		proxy: proxy,
		conf:  conf,
	}

	cc := &apps.Context{
		ContextFromUserAgent: apps.ContextFromUserAgent{
			TeamID: "some_team_id",
		},
	}

	bindings := []*apps.Binding{{Location: apps.LocationCommand}}

	proxy.EXPECT().CleanUserCallContext("some_user_id", cc).Return(cc, nil)
	proxy.EXPECT().GetBindings("some_session_id", "some_user_id", cc).Return(bindings, nil)
	conf.EXPECT().GetConfig().Return(config.Config{})

	res, err := api.handleGetBindings("some_session_id", "some_user_id", cc)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, bindings, res)
}
