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

	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type OAuth2Store struct{}

func (s *OAuth2Store) CreateState(r *incoming.Request) (string, error) {
	// fit the max key size of ~50chars
	buf := make([]byte, 15)
	_, _ = rand.Read(buf)
	state := fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString(buf), r.ActingUserID())
	_, err := r.API.Mattermost.KV.Set(OAuth2StatePrefix+state, state, pluginapi.SetExpiry(15*time.Minute))
	if err != nil {
		return "", err
	}
	return state, nil
}

func (s *OAuth2Store) ValidateStateOnce(r *incoming.Request, urlState string) error {
	ss := strings.Split(urlState, ".")
	if len(ss) != 2 || ss[1] != r.ActingUserID() {
		return utils.ErrForbidden
	}

	storedState := ""
	key := OAuth2StatePrefix + urlState
	err := r.API.Mattermost.KV.Get(key, &storedState)
	_ = r.API.Mattermost.KV.Delete(key)
	if err != nil {
		return err
	}
	if storedState != urlState {
		return utils.NewForbiddenError("state mismatch")
	}

	return nil
}

func (s *OAuth2Store) SaveUser(r *incoming.Request, data []byte) error {
	if r.SourceAppID() == "" || r.ActingUserID() == "" {
		return utils.NewInvalidError("app and user IDs must be provided")
	}

	userkey, err := Hashkey(UserPrefix, r.SourceAppID(), r.ActingUserID(), "", OAuth2UserKey)
	if err != nil {
		return err
	}

	_, err = r.API.Mattermost.KV.Set(userkey, data)
	return err
}

func (s *OAuth2Store) GetUser(r *incoming.Request) ([]byte, error) {
	if r.SourceAppID() == "" || r.ActingUserID() == "" {
		return nil, utils.NewInvalidError("app and user IDs must be provided")
	}

	userkey, err := Hashkey(UserPrefix, r.SourceAppID(), r.ActingUserID(), "", OAuth2UserKey)
	if err != nil {
		return nil, err
	}

	var data []byte
	if err = r.API.Mattermost.KV.Get(userkey, &data); err != nil {
		return nil, err
	}

	return data, nil
}
