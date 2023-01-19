package store

import (
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type SessionStore struct{}

type Sessions interface {
	Get(_ *incoming.Request, _ apps.AppID, userID string) (*model.Session, error)
	Save(_ *incoming.Request, _ apps.AppID, userID string, session *model.Session) error
	ListForUser(r *incoming.Request, userID string) ([]*model.Session, error)
	ListForApp(r *incoming.Request, appID apps.AppID) ([]*model.Session, error)
	Delete(r *incoming.Request, appID apps.AppID, userID string) error
}

func sessionKey(appID apps.AppID, userID string) string {
	return appKey(appID) + "_" + userID
}

func appKey(appID apps.AppID) string {
	return KVTokenPrefix + "_" + string(appID)
}

func parseSessionKey(key string) (apps.AppID, string, error) { //nolint:golint,unparam
	s := strings.Split(key, "_")
	if len(s) != 3 {
		return "", "", errors.New("invalid key pattern")
	}

	if s[0] != KVTokenPrefix {
		return "", "", errors.New("invalid key prefix")
	}

	return apps.AppID(s[1]), s[2], nil
}

func (s SessionStore) Get(r *incoming.Request, appID apps.AppID, userID string) (*model.Session, error) {
	return s.get(r, sessionKey(appID, userID))
}

func (s SessionStore) get(r *incoming.Request, key string) (*model.Session, error) {
	var session model.Session
	err := r.Config().MattermostAPI().KV.Get(key, &session)
	if err != nil {
		return nil, err
	}

	if session.Id == "" {
		return nil, utils.ErrNotFound
	}

	return &session, nil
}

func (s SessionStore) Save(r *incoming.Request, appID apps.AppID, userID string, session *model.Session) error {
	_, err := r.Config().MattermostAPI().KV.Set(sessionKey(appID, userID), session)
	if err != nil {
		return err
	}

	return nil
}

func (s SessionStore) listKeysForApp(r *incoming.Request, appID apps.AppID) ([]string, error) {
	ret := make([]string, 0)

	for i := 0; ; i++ {
		keys, err := r.Config().MattermostAPI().KV.ListKeys(i, ListKeysPerPage)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list keys - page, %d", i)
		}
		if len(keys) == 0 {
			break
		}
		for _, key := range keys {
			if strings.HasPrefix(key, appKey(appID)) {
				ret = append(ret, key)
			}
		}
	}

	return ret, nil
}

func (s SessionStore) listKeysForUser(r *incoming.Request, userID string) ([]string, error) {
	ret := make([]string, 0)

	for i := 0; ; i++ {
		keys, err := r.Config().MattermostAPI().KV.ListKeys(i, ListKeysPerPage)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list keys - page, %d", i)
		}
		if len(keys) == 0 {
			break
		}

		for _, key := range keys {
			_, keyUserID, err := parseSessionKey(key)
			if err != nil {
				continue
			}
			if keyUserID != userID {
				continue
			}

			ret = append(ret, key)
		}
	}

	return ret, nil
}

func (s SessionStore) ListForApp(r *incoming.Request, appID apps.AppID) ([]*model.Session, error) {
	keys, err := s.listKeysForApp(r, appID)
	if err != nil {
		return nil, err
	}

	ret := make([]*model.Session, 0)

	for _, key := range keys {
		session, err := s.get(r, key)
		if err != nil {
			return nil, errors.Wrapf(err, "failed get key, %s", key)
		}

		ret = append(ret, session)
	}

	return ret, nil
}

func (s SessionStore) ListForUser(r *incoming.Request, userID string) ([]*model.Session, error) {
	ret := make([]*model.Session, 0)

	for i := 0; ; i++ {
		keys, err := r.Config().MattermostAPI().KV.ListKeys(i, ListKeysPerPage)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list keys - page, %d", i)
		}
		if len(keys) == 0 {
			break
		}

		for _, key := range keys {
			_, keyUserID, err := parseSessionKey(key)
			if err != nil {
				continue
			}

			if keyUserID != userID {
				continue
			}

			session, err := s.get(r, key)
			if err != nil {
				r.Log.WithError(err).Debugf("failed get session for key, %s", key)
				continue
			}

			ret = append(ret, session)
		}
	}

	return ret, nil
}

func (s SessionStore) Delete(r *incoming.Request, appID apps.AppID, userID string) error {
	return r.Config().MattermostAPI().KV.Delete(sessionKey(appID, userID))
}

func (s SessionStore) DeleteAllForApp(r *incoming.Request, appID apps.AppID) error {
	keys, err := s.listKeysForApp(r, appID)
	if err != nil {
		return err
	}

	mm := r.Config().MattermostAPI()
	for _, key := range keys {
		err := mm.KV.Delete(key)
		if err != nil {
			r.Log.WithError(err).Debugf("failed delete session for key: %s, appID: %s", key, appID)
		}
	}

	return nil
}

func (s SessionStore) DeleteAllForUser(r *incoming.Request, userID string) error {
	keys, err := s.listKeysForUser(r, userID)
	if err != nil {
		return err
	}

	mm := r.Config().MattermostAPI()
	for _, key := range keys {
		err := mm.KV.Delete(key)
		if err != nil {
			r.Log.WithError(err).Debugf("failed delete session for key: %s, userID: %s", key, userID)
		}
	}

	return nil
}
