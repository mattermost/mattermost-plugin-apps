package store

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type SessionStore interface {
	Get(appID apps.AppID, userID string) (*model.Session, error)
	ListForApp(appID apps.AppID) ([]*model.Session, error)
	Save(appID apps.AppID, userID string, session *model.Session) error
	DeleteByID(sessionID string) error
}

type sessionStore struct {
	*Service
}

var _ SessionStore = (*sessionStore)(nil)

func sessionKey(appID apps.AppID, userID string) string {
	return appKey(appID) + "." + userID
}

func appKey(appID apps.AppID) string {
	return config.KVTokenPrefix + "." + string(appID)
}

func (s sessionStore) Get(appID apps.AppID, userID string) (*model.Session, error) {
	return s.get(sessionKey(appID, userID))
}

func (s sessionStore) get(key string) (*model.Session, error) {
	var session model.Session
	err := s.conf.MattermostAPI().KV.Get(key, &session)
	if err != nil {
		return nil, err
	}

	if session.Id == "" {
		return nil, utils.ErrNotFound
	}

	return &session, nil
}

func (s sessionStore) Save(appID apps.AppID, userID string, session *model.Session) error {
	_, err := s.conf.MattermostAPI().KV.Set(sessionKey(appID, userID), session)
	if err != nil {
		return err
	}

	return nil
}

func (s sessionStore) ListForApp(appID apps.AppID) ([]*model.Session, error) {
	ret := make([]*model.Session, 0)

	for i := 0; ; i++ {
		keys, err := s.conf.MattermostAPI().KV.ListKeys(i, keysPerPage, pluginapi.WithPrefix(appKey(appID)))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list keys - page, %d", i)
		}

		for _, k := range keys {
			session, err := s.get(k)
			if err != nil {
				return nil, errors.Wrapf(err, "failed get session for key, %s", k)
			}

			ret = append(ret, session)
		}

		if len(keys) < keysPerPage {
			break
		}
	}

	return ret, nil
}

func (s sessionStore) DeleteByID(sessionID string) error {
	/*
		err := s.conf.MattermostAPI().KV.Delete(key)
		if err != nil {
			return nil, err
		}
	*/
	return nil
}
