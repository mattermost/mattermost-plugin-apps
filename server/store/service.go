// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"encoding/ascii85"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

type Service struct {
	App          AppStore
	Subscription SubscriptionStore
	Manifest     ManifestStore
	AppKV        AppKVStore
	OAuth2       OAuth2Store

	mm   *pluginapi.Client
	conf config.Service
}

func MakeService(mm *pluginapi.Client, confService config.Service) (*Service, error) {
	s := &Service{
		mm:   mm,
		conf: confService,
	}
	s.AppKV = &appKVStore{
		Service: s,
	}
	s.OAuth2 = &oauth2Store{
		Service: s,
	}
	s.Subscription = &subscriptionStore{
		Service: s,
	}

	var err error
	s.App, err = makeAppStore(s)
	if err != nil {
		return nil, errors.New("failed to initialize App store")
	}

	s.Manifest, err = makeManifestStore(s)
	if err != nil {
		return nil, errors.New("failed to initialize App store")
	}
	return s, nil
}

func (s *Service) hashkey(globalNamespace, botUserID, appNamespace, key string) (string, error) {
	gns := []byte(globalNamespace)
	b := []byte(botUserID)
	k := []byte(key)

	ns := []byte(appNamespace)
	switch len(ns) {
	case 0:
		ns = []byte{' ', ' '}
	case 1:
		ns = []byte{ns[0], ' '}
	case 2:
		// nothing to do
	default:
		return "", errors.Errorf("prefix %q is longer than the limit of 2 ASCII characters", appNamespace)
	}

	switch {
	case len(k) == 0:
		return "", errors.New("key must not be empty")
	case len(b) != 26:
		return "", errors.Errorf("botUserID %q must be exactly 26 ASCII characters", botUserID)
	case len(gns) != 2 || gns[0] != '.':
		return "", errors.Errorf("global prefix %q is not 2 ASCII characters starting with a '.'", globalNamespace)
	}

	hashed := hashkey(gns, b, ns, k)
	if len(hashed) > model.KEY_VALUE_KEY_MAX_RUNES {
		return "", errors.Errorf("hashed key is too long (%v bytes), global namespace: %q, botUserID: %q, app namespace: %q, key: %q",
			len(hashed), globalNamespace, botUserID, appNamespace, key)
	}
	return hashed, nil
}

func hashkey(globalNamespace, botUserID, appNamespace, id []byte) string {
	idHash := make([]byte, 16)
	sha3.ShakeSum128(idHash, id)
	encodedID := make([]byte, ascii85.MaxEncodedLen(len(idHash)))
	_ = ascii85.Encode(encodedID, idHash)

	key := make([]byte, 0, model.KEY_VALUE_KEY_MAX_RUNES)
	key = append(key, globalNamespace...)
	key = append(key, botUserID...)
	key = append(key, appNamespace...)
	key = append(key, encodedID...)
	return string(key)
}

func parseHashkey(key string) (globalNamespace, botUserID, appNamespace, idhash string, err error) {
	k := []byte(key)
	if len(k) != model.KEY_VALUE_KEY_MAX_RUNES {
		return "", "", "", "", errors.Errorf("invalid key length %v bytes, must be %v", len(k), model.KEY_VALUE_KEY_MAX_RUNES)
	}
	gns := k[0:2]
	b := k[2:28]
	ns := k[28:30]
	h := k[30:50]

	return string(gns), string(b), strings.TrimSpace(string(ns)), string(h), nil
}
