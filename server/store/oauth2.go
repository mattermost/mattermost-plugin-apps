// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/pkg/errors"
)

type OAuth2Store interface {
	CreateState(actingUserID string) (string, error)
	ValidateStateOnce(urlState, actingUserID string) error
	SaveUser(appID apps.AppID, actingUserID string, data []byte) error
	GetUser(appID apps.AppID, actingUserID string) ([]byte, error)
	ClearUsers(appID apps.AppID) error
}

type oauth2Store struct {
	*Service
}

var _ OAuth2Store = (*oauth2Store)(nil)

func (s *oauth2Store) CreateState(actingUserID string) (string, error) {
	// fit the max key size of ~50chars
	buf := make([]byte, 15)
	_, _ = rand.Read(buf)
	state := fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString(buf), actingUserID)
	_, err := s.conf.MattermostAPI().KV.Set(KVOAuth2StatePrefix+state, state, pluginapi.SetExpiry(15*time.Minute))
	if err != nil {
		return "", err
	}
	return state, nil
}

func (s *oauth2Store) ValidateStateOnce(urlState, actingUserID string) error {
	ss := strings.Split(urlState, ".")
	if len(ss) != 2 || ss[1] != actingUserID {
		return utils.ErrForbidden
	}

	storedState := ""
	key := KVOAuth2StatePrefix + urlState
	err := s.conf.MattermostAPI().KV.Get(key, &storedState)
	_ = s.conf.MattermostAPI().KV.Delete(key)
	if err != nil {
		return err
	}
	if storedState != urlState {
		return utils.NewForbiddenError("state mismatch")
	}

	return nil
}

func (s *oauth2Store) SaveUser(appID apps.AppID, actingUserID string, data []byte) error {
	if appID == "" || actingUserID == "" {
		return utils.NewInvalidError("app and user IDs must be provided")
	}

	userkey, err := Hashkey(KVUserPrefix, appID, actingUserID, "", KVUserKey)
	if err != nil {
		return err
	}

	_, err = s.conf.MattermostAPI().KV.Set(userkey, data)
	return err
}

func (s *oauth2Store) GetUser(appID apps.AppID, actingUserID string) ([]byte, error) {
	if appID == "" || actingUserID == "" {
		return nil, utils.NewInvalidError("app and user IDs must be provided")
	}

	userkey, err := Hashkey(KVUserPrefix, appID, actingUserID, "", KVUserKey)
	if err != nil {
		return nil, err
	}

	var data []byte
	if err = s.conf.MattermostAPI().KV.Get(userkey, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func (s *oauth2Store) ClearUsers(appID apps.AppID) error {
	for i := 0; ; i++ {
		keys, err := s.conf.MattermostAPI().KV.ListKeys(i, ListKeysPerPage)
		if err != nil {
			return errors.Wrapf(err, "failed to list keys - page, %d", i)
		}

		if len(keys) == 0 {
			break
		}

		// Clear users
		for _, key := range keys {
			userAppID, keyUserID, err := parseSessionKey(key)
			if err != nil {
				continue
			}

			if userAppID == appID {
				s.SaveUser(appID, keyUserID, nil)
			}
		}
	}
	return nil
}
