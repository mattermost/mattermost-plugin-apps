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
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

type OAuth2Store interface {
	CreateState(actingUserID string) (string, error)
	ValidateStateOnce(urlState, actingUserID string) error
	SaveUser(_ apps.AppID, mattermostUserID string, ref interface{}) error
	GetUser(_ apps.AppID, mattermostUserID string, ref interface{}) error
}

type oauth2Store struct {
	*Service
}

var _ OAuth2Store = (*oauth2Store)(nil)

func (s *oauth2Store) CreateState(actingUserID string) (string, error) {
	// fit the max key size of ~50chars
	r := make([]byte, 15)
	_, _ = rand.Read(r)
	state := fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString(r), actingUserID)
	_, err := s.mm.KV.Set(config.KVOAuth2StatePrefix+state, state, pluginapi.SetExpiry(15*time.Minute))
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
	key := config.KVOAuth2StatePrefix + urlState
	err := s.mm.KV.Get(key, &storedState)
	_ = s.mm.KV.Delete(key)
	if err != nil {
		return err
	}
	if storedState != urlState {
		return utils.NewForbiddenError("state mismatch")
	}

	return nil
}

func (s *oauth2Store) SaveUser(appID apps.AppID, mattermostUserID string, ref interface{}) error {
	if appID == "" || mattermostUserID == "" {
		return utils.NewInvalidError("namespace and mattermost user ID must be provided")
	}
	_, err := s.mm.KV.Set(s.userKey(string(appID), mattermostUserID), ref)
	return err
}

func (s *oauth2Store) GetUser(appID apps.AppID, mattermostUserID string, ref interface{}) error {
	return s.mm.KV.Get(s.userKey(string(appID), mattermostUserID), ref)
}

func (s *oauth2Store) userKey(namespace, mattermostUserID string) string {
	return s.hashkey(config.KVOAuth2Prefix, namespace, "remote_user", mattermostUserID)
}
