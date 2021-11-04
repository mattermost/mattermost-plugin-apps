package store

import (
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type SessionStore interface {
	Get(appID apps.AppID, userID string) (*model.Session, error)
	Save(appID apps.AppID, userID string, session *model.Session) error
}

type sessionStore struct {
	*Service
}

var _ SessionStore = (*sessionStore)(nil)

func (s sessionStore) Get(appID apps.AppID, userID string) (*model.Session, error) {
	var session model.Session
	err := s.conf.MattermostAPI().KV.Get(tokenKey(appID, userID), &session)
	if err != nil {
		return nil, err
	}

	if session.Id == "" {
		return nil, utils.ErrNotFound
	}

	return &session, nil
}

func (s sessionStore) Save(appID apps.AppID, userID string, session *model.Session) error {
	_, err := s.conf.MattermostAPI().KV.Set(tokenKey(appID, userID), session)
	if err != nil {
		return err
	}

	return nil
}

func tokenKey(appID apps.AppID, userID string) string {
	return config.KVTokenPrefix + "." + string(appID) + "." + userID
}
