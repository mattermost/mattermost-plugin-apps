// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

type OAuth2Store interface {
	CreateState(actingUserID string) (string, error)
	GetStateOnce(urlState string) (string, error)
	SaveRemoteUser(namespace, mattermostUserID string, ref interface{}) error
	GetRemoteUser(namespace, mattermostUserID string, ref interface{}) error
}

type oauth2Store struct {
	*Service
}

var _ OAuth2Store = (*oauth2Store)(nil)

func (s *oauth2Store) CreateState(actingUserID string) (string, error) {
	// fit the max key size of ~50chars
	r := make([]byte, 15)
	rand.Read(r)
	state := fmt.Sprintf("state_%v_%s", base64.RawURLEncoding.EncodeToString(r), actingUserID)

	_, err := s.mm.KV.Set(config.KVOAuth2Prefix+state, state, pluginapi.SetExpiry(15*time.Minute))
	if err != nil {
		return "", err
	}

	return state, nil
}

func (s *oauth2Store) GetStateOnce(urlState string) (string, error) {
	storedState := ""
	key := config.KVOAuth2Prefix + urlState
	err := s.mm.KV.Get(key, &storedState)
	s.mm.KV.Delete(key)
	return storedState, err
}

func (s *oauth2Store) SaveRemoteUser(namespace, mattermostUserID string, ref interface{}) error {
	if namespace == "" || mattermostUserID == "" {
		return errors.New("namespace and mattermost user ID must be provided")
	}
	_, err := s.mm.KV.Set(remoteUserKey(namespace, mattermostUserID), ref)
	return err
}

func (s *oauth2Store) GetRemoteUser(namespace, mattermostUserID string, ref interface{}) error {
	return s.mm.KV.Get(remoteUserKey(namespace, mattermostUserID), ref)
}

func remoteUserKey(namespace, mattermostUserID string) string {
	return config.KVOAuth2Prefix + kvKey(namespace, "remote_user", mattermostUserID)
}
