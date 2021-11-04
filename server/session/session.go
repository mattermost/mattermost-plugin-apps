package session

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Service interface {
	GetOrCreate(appID apps.AppID, userID string) (*model.Session, error)
}

var _ Service = (*service)(nil)

type service struct {
	log   utils.Logger
	mm    *pluginapi.Client
	store store.Service
}

func NewService(mm *pluginapi.Client, store store.Service) Service {
	return &service{
		log: utils.NewPluginLogger(mm),
		mm:  mm,
	}
}

const (
	sessionLengthInMinutes = 10
)

func (s *service) GetOrCreate(appID apps.AppID, userID string) (*model.Session, error) {
	session, err := s.store.Session.Get(appID, userID)
	if err == nil {
		return session, nil
	}

	if !errors.Is(err, utils.ErrNotFound) {
		return nil, errors.Wrap(err, "failed to get session from store")
	}

	user, err := s.mm.User.Get(userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch user for new session")
	}

	app, err := s.store.App.Get(appID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch app for new session")
	}

	oAuthApp := app.MattermostOAuth2
	if oAuthApp == nil {
		return nil, errors.New("can't create session for app without Mattermost OAuth2 App")
	}

	session = &model.Session{
		UserId:    userID,
		Roles:     user.Roles,
		IsOAuth:   true,
		ExpiresAt: model.GetMillis() + (1000 * 60 * sessionLengthInMinutes),
	}
	session.GenerateCSRF()
	session.AddProp(model.SessionPropPlatform, oAuthApp.Name)
	session.AddProp(model.SessionPropOAuthAppID, oAuthApp.Id)
	session.AddProp(model.SessionPropAppsFrameworkAppID, oAuthApp.AppsFrameworkAppID)
	session.AddProp(model.SessionPropOs, "OAuth2")
	session.AddProp(model.SessionPropBrowser, "OAuth2")

	session, err = s.mm.Session.Create(session)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new app session")
	}

	err = s.store.Session.Save(appID, userID, session)
	if err != nil {
		return nil, errors.Wrap(err, "failed to save new session in store")
	}

	return session, nil
}
