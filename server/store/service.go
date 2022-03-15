// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"encoding/ascii85"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Service struct {
	App          AppStore
	Subscription SubscriptionStore
	Manifest     ManifestStore
	AppKV        AppKVStore
	OAuth2       OAuth2Store
	Session      SessionStore

	conf    config.Service
	httpOut httpout.Service
}

func MakeService(log utils.Logger, confService config.Service, httpOut httpout.Service) (*Service, error) {
	s := &Service{
		conf:    confService,
		httpOut: httpOut,
	}
	s.AppKV = &appKVStore{Service: s}
	s.OAuth2 = &oauth2Store{Service: s}
	s.Subscription = &subscriptionStore{Service: s}
	s.Session = &sessionStore{Service: s}

	conf := confService.Get()
	var err error
	s.App, err = makeAppStore(s, conf, log)
	if err != nil {
		return nil, err
	}

	s.Manifest, err = makeManifestStore(s, conf, log)
	if err != nil {
		return nil, err
	}
	return s, nil
}

const (
	hashKeyLength = 82
)

func Hashkey(globalNamespace string, appID apps.AppID, userID, appNamespace, key string) (string, error) {
	gns := []byte(globalNamespace)
	a := []byte(appID)

	for len(a) < apps.MaxAppIDLength {
		a = append(a, ' ')
	}

	u := []byte(userID)
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
	case len(a) != 32:
		return "", errors.New("appID must be max length")
	case len(u) != 26:
		return "", errors.Errorf("userID %q must be exactly 26 ASCII characters", userID)
	case len(gns) != 2 || gns[0] != '.':
		return "", errors.Errorf("global prefix %q is not 2 ASCII characters starting with a '.'", globalNamespace)
	}

	hashed := hashkey(gns, a, u, ns, k)
	if len(hashed) != hashKeyLength {
		return "", errors.Errorf("hashed key has wrong length (%v bytes), global namespace: %q, appID: %q, userID: %q, app namespace: %q, key: %q",
			len(hashed), globalNamespace, appID, userID, appNamespace, key)
	}
	return hashed, nil
}

func hashkey(globalNamespace, appID, userID, appNamespace, id []byte) string {
	idHash := make([]byte, 16)
	sha3.ShakeSum128(idHash, id)
	encodedID := make([]byte, ascii85.MaxEncodedLen(len(idHash)))
	_ = ascii85.Encode(encodedID, idHash)

	key := make([]byte, 0, model.KeyValueKeyMaxRunes)
	key = append(key, globalNamespace...)
	key = append(key, appID...)
	key = append(key, userID...)
	key = append(key, appNamespace...)
	key = append(key, encodedID...)
	return string(key)
}

func ParseHashkey(key string) (globalNamespace string, appID apps.AppID, userID, appNamespace, idhash string, err error) {
	k := []byte(key)
	if len(k) != hashKeyLength {
		return "", "", "", "", "", errors.Errorf("invalid key length %v bytes, must be smaller then %v", len(k), model.KeyValueKeyMaxRunes)
	}
	gns := k[0:2]
	a := k[2:34]
	u := k[34:60]
	ns := k[60:62]
	h := k[62:82]

	return string(gns), apps.AppID(strings.TrimSpace(string(a))), string(u), strings.TrimSpace(string(ns)), string(h), nil
}
