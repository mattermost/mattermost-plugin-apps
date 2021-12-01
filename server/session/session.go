package session

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/sessionutils"
)

const (
	sessionLengthInMinutes    = 10
	minSessionLengthInMinutes = 5
)

type Service interface {
	GetOrCreate(r *incoming.Request, appID apps.AppID, userID string) (*model.Session, error)
	ListForUser(r *incoming.Request, userID string) ([]*model.Session, error)
	RevokeSessionsForApp(r *incoming.Request, appID apps.AppID) error
}

var _ Service = (*service)(nil)

type service struct {
	log   utils.Logger
	mm    *pluginapi.Client
	store *store.Service
}

func NewService(mm *pluginapi.Client, store *store.Service) Service {
	return &service{
		log:   utils.NewPluginLogger(mm),
		mm:    mm,
		store: store,
	}
}

func (s *service) GetOrCreate(r *incoming.Request, appID apps.AppID, userID string) (*model.Session, error) {
	session, err := s.store.Session.Get(r, appID, userID)

	if err == nil && !session.IsExpired() {
		err = s.extendSessionExpiryIfNeeded(r, appID, userID, session)
		if err != nil {
			return nil, errors.Wrap(err, "failed to extend session length")
		}

		return session, nil
	}

	if err != nil && !errors.Is(err, utils.ErrNotFound) {
		return nil, errors.Wrap(err, "failed to get session from store")
	}

	return s.createSession(r, appID, userID)
}

func (s *service) createSession(r *incoming.Request, appID apps.AppID, userID string) (*model.Session, error) {
	user, err := s.mm.User.Get(userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch user for new session")
	}

	app, err := s.store.App.Get(r, appID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch app for new session")
	}

	if app.DeployType == apps.DeployBuiltin {
		return nil, errors.New("builtin apps can't have app specific session")
	}

	session := &model.Session{
		UserId:    userID,
		Roles:     user.Roles,
		IsOAuth:   true,
		ExpiresAt: model.GetMillis() + (1000 * 60 * sessionLengthInMinutes),
	}
	session.GenerateCSRF()

	oAuthApp := app.MattermostOAuth2

	session.AddProp(model.SessionPropOs, "OAuth2")
	session.AddProp(model.SessionPropBrowser, "OAuth2")
	session.AddProp(model.SessionPropPlatform, oAuthApp.Name)
	session.AddProp(model.SessionPropOAuthAppID, oAuthApp.Id)
	session.AddProp(model.SessionPropAppsFrameworkAppID, oAuthApp.AppsFrameworkAppID)

	session, err = s.mm.Session.Create(session)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new app session")
	}

	err = s.store.Session.Save(r, appID, userID, session)
	if err != nil {
		return nil, errors.Wrap(err, "failed to save new session in store")
	}

	s.log.Debugw("Created new access token", "app_id", appID, "user_id", userID, "token", utils.LastN(session.Token, 3))

	return session, nil
}

func (s *service) extendSessionExpiryIfNeeded(r *incoming.Request, appID apps.AppID, userID string, session *model.Session) error {
	minSessionLength := int64(1000 * 60 * minSessionLengthInMinutes)

	now := model.GetMillis()
	remaining := session.ExpiresAt - now
	if remaining > minSessionLength {
		return nil
	}

	newExpiryTime := now + (1000 * 60 * sessionLengthInMinutes)

	err := s.mm.Session.ExtendExpiry(session.Id, newExpiryTime)
	if err != nil {
		return err
	}

	// Update store
	session.ExpiresAt = newExpiryTime

	err = s.store.Session.Save(r, appID, userID, session)
	if err != nil {
		return errors.Wrap(err, "failed to save new session in store")
	}

	return nil
}

func (s service) ListForUser(r *incoming.Request, userID string) ([]*model.Session, error) {
	return s.store.Session.ListForUser(r, userID)
}

func (s service) RevokeSessionsForApp(r *incoming.Request, appID apps.AppID) error {
	sessions, err := s.store.Session.ListForApp(r, appID)
	if err != nil {
		return errors.Wrap(err, "failed to list app sessions for revocation")
	}

	for _, session := range sessions {
		// Revoke active sessions
		if !session.IsExpired() {
			err = s.mm.Session.Revoke(session.Id)
			if err != nil {
				r.Log.WithError(err).Warnw("failed to revoke app session")
			}
		}

		err = s.store.Session.Delete(r, sessionutils.GetAppID(session), session.UserId)
		if err != nil {
			r.Log.WithError(err).Warnw("failed to delete revoked session from store")
		}

		r.Log.Warnf("revoked session: %#+v\n", session.Id)
	}

	return nil
}
