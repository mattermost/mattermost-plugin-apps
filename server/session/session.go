package session

import (
	"log"
	"math"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const (
	sessionLengthInMinutes = 10
)

type Service interface {
	GetOrCreate(appID apps.AppID, userID string) (*model.Session, error)
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

func (s *service) GetOrCreate(appID apps.AppID, userID string) (*model.Session, error) {
	s.log.Warnw("GetOrCreate", "userID", userID, "appID", string(appID))
	session, err := s.store.Session.Get(appID, userID)
	log.Printf("session: %#+v\n", session)
	log.Printf("err: %#+v\n", err)

	if err == nil && !session.IsExpired() {
		log.Println("in if")
		err = s.extendSessionExpiryIfNeeded(session)
		if err != nil {
			return nil, errors.Wrap(err, "failed to extend session length")
		}

		s.log.Warnw("found access token", "token", session.Token)

		return session, nil
	}

	log.Println("not in if")

	log.Printf("session1: %#+v\n", session)

	if err != nil && !errors.Is(err, utils.ErrNotFound) {
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

	session = &model.Session{
		UserId:    userID,
		Roles:     user.Roles,
		IsOAuth:   true,
		ExpiresAt: model.GetMillis() + (1000 * 60 * sessionLengthInMinutes),
	}
	session.GenerateCSRF()

	oAuthApp := app.MattermostOAuth2
	if oAuthApp != nil {
		// TODO: The buit-in app also needs some (?) of this props. Does it also need an OAuth app?
		session.AddProp(model.SessionPropPlatform, oAuthApp.Name)
		session.AddProp(model.SessionPropOAuthAppID, oAuthApp.Id)
		session.AddProp(model.SessionPropAppsFrameworkAppID, oAuthApp.AppsFrameworkAppID)
	}

	session.AddProp(model.SessionPropOs, "OAuth2")
	session.AddProp(model.SessionPropBrowser, "OAuth2")

	log.Printf("session2: %#+v\n", session)

	session, err = s.mm.Session.Create(session)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new app session")
	}

	log.Printf("session3: %#v\n", session)

	err = s.store.Session.Save(appID, userID, session)
	if err != nil {
		return nil, errors.Wrap(err, "failed to save new session in store")
	}

	log.Printf("session4: %#v\n", session)

	s.log.Warnw("created new access token", "token", session.Token)

	return session, nil
}

func (s *service) extendSessionExpiryIfNeeded(session *model.Session) error {
	return nil

	sessionLength := int64(1000 * 60 * sessionLengthInMinutes)

	// Only extend the expiry if the lessor of 1% or 1 day has elapsed within the
	// current session duration.
	threshold := int64(math.Min(float64(sessionLength)*0.01, float64(24*60*60*1000)))
	// Minimum session length is 1 day as of this writing, therefore a minimum ~14 minutes threshold.
	// However we'll add a sanity check here in case that changes. Minimum 5 minute threshold,
	// meaning we won't write a new expiry more than every 5 minutes.
	if threshold < 5*60*1000 {
		threshold = 5 * 60 * 1000
	}

	now := model.GetMillis()
	elapsed := now - (session.ExpiresAt - sessionLength)
	if elapsed < threshold {
		return nil
	}

	err := s.mm.Session.ExtendExpiry(session.Id, now) // TODO
	if err != nil {
		return err
	}

	return nil
}
