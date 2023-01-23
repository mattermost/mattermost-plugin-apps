package session_test

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest/mock"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_store"
	"github.com/mattermost/mattermost-plugin-apps/server/session"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/sessionutils"
)

func setUpBasics(ctrl *gomock.Controller) (session.Service,
	*incoming.Request,
	*mock_store.MockSessionStore,
	*mock_store.MockAppStore,
	*plugintest.API) {
	appStore := mock_store.NewMockAppStore(ctrl)
	sessionStore := mock_store.NewMockSessionStore(ctrl)
	mockStore := &store.Service{
		App:     appStore,
		Session: sessionStore,
	}

	conf, api := config.NewTestService(nil)
	r := incoming.NewRequest(conf, nil)

	sessionService := session.NewService(conf.MattermostAPI(), mockStore)

	return sessionService, r, sessionStore, appStore, api
}

func TestGetOrCreate(t *testing.T) {
	t.Parallel()

	t.Run("Valid, long lasting session found", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		sessionService, r, sessionStore, _, _ := setUpBasics(ctrl)

		appID := apps.AppID("foo")
		userID := model.NewId()

		expires := time.Now().Add(session.SessionLength)
		session := &model.Session{
			Id:        model.NewId(),
			UserId:    userID,
			ExpiresAt: expires.UnixMilli(),
			IsOAuth:   true,
		}

		session.AddProp(model.SessionPropMattermostAppID, string(appID))
		sessionStore.EXPECT().Get(appID, userID).Times(1).Return(session, nil)

		r = r.WithDestination(appID)
		rSession, err := sessionService.GetOrCreate(r, userID)
		assert.NoError(t, err)
		assert.NotNil(t, rSession)
	})

	t.Run("Valid, short session found", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		sessionService, r, sessionStore, _, api := setUpBasics(ctrl)

		appID := apps.AppID("foo")
		userID := model.NewId()

		expires := time.Now().Add(session.MinSessionLength - 1*time.Minute) // 4 Minutes left
		s := model.Session{
			Id:        model.NewId(),
			UserId:    userID,
			ExpiresAt: expires.UnixMilli(),
			IsOAuth:   true,
		}

		api.On("ExtendSessionExpiry", s.Id, mock.Anything).Once().Return(nil)

		s.AddProp(model.SessionPropMattermostAppID, string(appID))
		sessionStore.EXPECT().Get(appID, userID).Times(1).Return(&s, nil)

		sessionStore.EXPECT().Save(appID, userID, gomock.Any()).Times(1).Return(nil)

		r = r.WithDestination(appID)
		rSession, err := sessionService.GetOrCreate(r, userID)
		assert.NoError(t, err)
		require.NotNil(t, rSession)
		assert.Equal(t, userID, rSession.UserId)
		assert.True(t, rSession.IsOAuth)
		assert.Equal(t, appID, sessionutils.GetAppID(rSession))
		// Assert that the new session has at least 9 minutes left
		minExpiresAt := time.Now().Add(session.SessionLength - time.Minute).UnixMilli()
		assert.GreaterOrEqual(t, rSession.ExpiresAt, minExpiresAt)
	})

	t.Run("No session found", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		sessionService, r, sessionStore, appStore, api := setUpBasics(ctrl)

		appID := apps.AppID("foo")
		userID := model.NewId()

		api.On("GetUser", userID).Return(&model.User{
			Id:    userID,
			Roles: "",
		}, nil)

		oAuthApp := &model.OAuthApp{
			Id:   model.NewId(),
			Name: "App Name",
		}
		newSession := &model.Session{
			UserId:  userID,
			IsOAuth: true,
		}

		newSession.AddProp(model.SessionPropOs, "OAuth2")
		newSession.AddProp(model.SessionPropBrowser, "OAuth2")
		newSession.AddProp(model.SessionPropPlatform, oAuthApp.Name)
		newSession.AddProp(model.SessionPropOAuthAppID, oAuthApp.Id)
		newSession.AddProp(model.SessionPropMattermostAppID, string(appID))

		api.On("CreateSession", mock.AnythingOfType("*model.Session")).Run(func(args mock.Arguments) {
			rSession, ok := args[0].(*model.Session)
			require.True(t, ok)
			require.NotNil(t, rSession)

			// Test if new session expires in at least 9 minutes
			assert.GreaterOrEqual(t, rSession.ExpiresAt, time.Now().Add(session.SessionLength-time.Minute).UnixMilli())

			// Copy over field that can't be asserted
			newSession.ExpiresAt = rSession.ExpiresAt
			newSession.AddProp("csrf", rSession.Props["csrf"])
			assert.Equal(t, newSession, rSession)
		}).Return(newSession, nil)

		sessionStore.EXPECT().Get(appID, userID).Times(1).Return(nil, utils.ErrNotFound)
		sessionStore.EXPECT().Save(appID, userID, gomock.Any()).Times(1).Return(nil)
		appStore.EXPECT().Get(appID, store.EnabledAppsOnly).Times(1).Return(&apps.App{
			MattermostOAuth2: oAuthApp,
		}, nil)

		r = r.WithDestination(appID)
		rSession, err := sessionService.GetOrCreate(r, userID)
		assert.NoError(t, err)
		assert.NotNil(t, rSession)
	})
}

func TestListForUser(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	sessionService, r, sessionStore, _, _ := setUpBasics(ctrl)

	userID := model.NewId()
	sessionStore.EXPECT().ListForUser(r, userID).Times(1).Return([]*model.Session{}, nil)

	rSessions, err := sessionService.ListForUser(r, userID)
	assert.NoError(t, err)
	assert.Equal(t, []*model.Session{}, rSessions)
}

func TestRevokeSessionsForApp(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	sessionService, r, sessionStore, _, api := setUpBasics(ctrl)

	appID := apps.AppID("foo")
	userID1 := model.NewId()
	userID2 := model.NewId()

	session1 := &model.Session{
		Id:      model.NewId(),
		UserId:  userID1,
		IsOAuth: true,
	}
	session1.AddProp(model.SessionPropMattermostAppID, string(appID))
	session2 := &model.Session{
		Id:      model.NewId(),
		UserId:  userID2,
		IsOAuth: true,
	}
	session2.AddProp(model.SessionPropMattermostAppID, string(appID))
	sessions := []*model.Session{session1, session2}

	sessionStore.EXPECT().ListForApp(appID).Return(sessions, nil).Times(1)
	sessionStore.EXPECT().Delete(appID, userID1).Return(nil).Times(1)
	sessionStore.EXPECT().Delete(appID, userID2).Return(nil).Times(1)

	api.On("RevokeSession", sessions[0].Id).Return(nil).Once()
	api.On("RevokeSession", sessions[1].Id).Return(nil).Once()

	err := sessionService.RevokeSessionsForApp(r, appID)
	assert.NoError(t, err)
}
