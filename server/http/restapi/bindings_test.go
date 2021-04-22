package restapi

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_proxy"
)

func TestHandleGetBindingsInvalidContext(t *testing.T) {
	ctrl := gomock.NewController(t)

	proxy := mock_proxy.NewMockService(ctrl)
	conf := mock_config.NewMockService(ctrl)

	testAPI := &plugintest.API{}
	testAPI.On("LogDebug", mock.Anything).Return(nil)
	mm := pluginapi.NewClient(testAPI)

	a := &restapi{
		proxy: proxy,
		conf:  conf,
		mm:    mm,
	}

	cc := &apps.Context{
		ContextFromUserAgent: apps.ContextFromUserAgent{
			TeamID: "some_team_id",
		},
	}

	testAPI.On("GetTeamMember", "some_team_id", "some_user_id").Return(nil, &model.AppError{
		Message: "user is not a member of the specified team",
	})

	res, err := a.handleGetBindings("some_session_id", "some_user_id", cc)
	require.Error(t, err)
	require.Nil(t, res)
}

func TestHandleGetBindingsValidContext(t *testing.T) {
	ctrl := gomock.NewController(t)

	proxy := mock_proxy.NewMockService(ctrl)
	conf := mock_config.NewMockService(ctrl)

	testAPI := &plugintest.API{}
	testAPI.On("LogDebug", mock.Anything).Return(nil)
	mm := pluginapi.NewClient(testAPI)

	api := &restapi{
		proxy: proxy,
		conf:  conf,
		mm:    mm,
	}

	cc := &apps.Context{
		ContextFromUserAgent: apps.ContextFromUserAgent{
			TeamID: "some_team_id",
		},
	}

	bindings := []*apps.Binding{{Location: apps.LocationCommand}}

	testAPI.On("GetTeamMember", "some_team_id", "some_user_id").Return(&model.TeamMember{
		TeamId: "some_team_id",
		UserId: "some_user_id",
	}, nil)

	proxy.EXPECT().GetBindings("some_session_id", "some_user_id", cc).Return(bindings, nil)
	conf.EXPECT().GetConfig().Return(config.Config{})

	res, err := api.handleGetBindings("some_session_id", "some_user_id", cc)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, bindings, res)
}
