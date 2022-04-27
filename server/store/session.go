package store

import (
	"strings"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type SessionStore interface {
	Get(_ apps.AppID, userID string) (*model.Session, error)
	ListForApp(apps.AppID) ([]*model.Session, error)
	ListForUser(_ *incoming.Request, userID string) ([]*model.Session, error)
	Save(_ apps.AppID, userID string, session *model.Session) error
	Delete(_ apps.AppID, userID string) error
	DeleteAllForApp(*incoming.Request, apps.AppID) error
	DeleteAllForUser(_ *incoming.Request, userID string) error
}

type sessionStore struct {
	*Service
}

var _ SessionStore = (*sessionStore)(nil)

func sessionKey(appID apps.AppID, userID string) string {
	return appKey(appID) + "_" + userID
}

func appKey(appID apps.AppID) string {
	return config.KVTokenPrefix + "_" + string(appID)
}

func parseKey(key string) (apps.AppID, string, error) {
	s := strings.Split(key, "_")
	if len(s) != 3 {
		return "", "", errors.New("invalid key pattern")
	}

	if s[0] != config.KVTokenPrefix {
		return "", "", errors.New("invalid key prefix")
	}

	return apps.AppID(s[1]), s[2], nil
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

func (s sessionStore) listKeysForApp(appID apps.AppID) ([]string, error) {
	ret := make([]string, 0)

	for i := 0; ; i++ {
		keys, err := s.conf.MattermostAPI().KV.ListKeys(i, keysPerPage, pluginapi.WithPrefix(appKey(appID)))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list keys - page, %d", i)
		}

		ret = append(ret, keys...)

		if len(keys) < keysPerPage {
			break
		}
	}

	return ret, nil
}

func (s sessionStore) listKeysForUser(userID string) ([]string, error) {
	ret := make([]string, 0)

	for i := 0; ; i++ {
		keys, err := s.conf.MattermostAPI().KV.ListKeys(i, keysPerPage)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list keys - page, %d", i)
		}

		for _, key := range keys {
			_, keyUserID, err := parseKey(key)
			if err != nil {
				continue
			}

			if keyUserID != userID {
				continue
			}

			ret = append(ret, key)
		}

		if len(keys) < keysPerPage {
			break
		}
	}

	return ret, nil
}

func (s sessionStore) ListForApp(appID apps.AppID) ([]*model.Session, error) {
	keys, err := s.listKeysForApp(appID)
	if err != nil {
		return nil, err
	}

	ret := make([]*model.Session, 0)

	for _, key := range keys {
		session, err := s.get(key)
		if err != nil {
			return nil, errors.Wrapf(err, "failed get key, %s", key)
		}

		ret = append(ret, session)
	}

	return ret, nil
}

func (s sessionStore) ListForUser(r *incoming.Request, userID string) ([]*model.Session, error) {
	ret := make([]*model.Session, 0)

	for i := 0; ; i++ {
		keys, err := s.conf.MattermostAPI().KV.ListKeys(i, keysPerPage)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list keys - page, %d", i)
		}

		for _, key := range keys {
			_, keyUserID, err := parseKey(key)
			if err != nil {
				continue
			}

			if keyUserID != userID {
				continue
			}

			session, err := s.get(key)
			if err != nil {
				r.Log.WithError(err).Debugf("failed get session for key, %s", key)
				continue
			}

			ret = append(ret, session)
		}

		if len(keys) < keysPerPage {
			break
		}
	}

	return ret, nil
}

func (s sessionStore) Delete(appID apps.AppID, userID string) error {
	return s.conf.MattermostAPI().KV.Delete(sessionKey(appID, userID))
}

func (s sessionStore) DeleteAllForApp(r *incoming.Request, appID apps.AppID) error {
	keys, err := s.listKeysForApp(appID)
	if err != nil {
		return err
	}

	for _, key := range keys {
		err := s.conf.MattermostAPI().KV.Delete(key)
		if err != nil {
			r.Log.WithError(err).Debugf("failed delete session for key: %s, appID: %s", key, appID)
		}
	}

	return nil
}

func (s sessionStore) DeleteAllForUser(r *incoming.Request, userID string) error {
	keys, err := s.listKeysForUser(userID)
	if err != nil {
		return err
	}

	for _, key := range keys {
		err := s.conf.MattermostAPI().KV.Delete(key)
		if err != nil {
			r.Log.WithError(err).Debugf("failed delete session for key: %s, userID: %s", key, userID)
		}
	}

	return nil
}
