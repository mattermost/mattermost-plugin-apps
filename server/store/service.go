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
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

// KV namespace
//
// Keys starting with a '.' are reserved for app-specific keys in the "hashkey"
// format. Hashkeys have the following format (see service_test.go#TestHashkey
// for examples):
//
//   - global prefix of ".X" where X is exactly 1 byte (2 bytes)
//   - bot user ID (26 bytes)
//   - app-specific prefix, limited to 2 non-space ASCII characters, right-filled
//     with ' ' to 2 bytes.
//   - app-specific key hash: 16 bytes, ascii85 (20 bytes)
//
// All other keys must start with an ASCII letter. '.' is usually used as the
// terminator since it is not used in the base64 representation.
const (
	// KVAppPrefix is the Apps global namespace.
	KVAppPrefix = ".k"

	// KVUserPrefix is the global namespace used to store user
	// records.
	KVUserPrefix = ".u"

	// KVUserPrefix is the key to store OAuth2 user
	// records.
	KVUserKey = "oauth2_user"

	// KVOAuth2StatePrefix is the global namespace used to store OAuth2
	// ephemeral state data.
	KVOAuth2StatePrefix = ".o"

	// KVSubPrefix is used for keys storing subscriptions.
	KVSubPrefix = "sub."

	// KVInstalledAppPrefix is used to store App records.
	KVInstalledAppPrefix = "app."

	// KVLocalManifestPrefix is used to store locally-listed manifests.
	KVLocalManifestPrefix = "man."

	KVTokenPrefix = ".t"

	KVDebugPrefix = ".debug."

	// KVCallOnceKey and KVClusterMutexKey are used for invoking App Calls once,
	// usually upon a Mattermost instance startup.
	KVCallOnceKey     = "CallOnce"
	KVClusterMutexKey = "Cluster_Mutex"
)

const (
	ListKeysPerPage = 1000
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

func MakeService(confService config.Service, httpOut httpout.Service) (*Service, error) {
	s := &Service{
		conf:    confService,
		httpOut: httpOut,
	}
	s.AppKV = &appKVStore{Service: s}
	s.OAuth2 = &oauth2Store{
		Service:   s,
		encrypter: &AESEncrypter{},
	}
	s.Subscription = &subscriptionStore{Service: s}
	s.Session = &sessionStore{Service: s}

	conf := confService.Get()
	var err error
	s.App, err = s.makeAppStore(conf)
	if err != nil {
		return nil, err
	}

	s.Manifest, err = s.makeManifestStore(conf)
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

func (s *Service) ListHashKeys(
	r *incoming.Request,
	processf func(key string) error,
	matchf ...func(prefix string, _ apps.AppID, userID, namespace, idhash string) bool,
) error {
	mm := s.conf.MattermostAPI()
	for pageNumber := 0; ; pageNumber++ {
		keys, err := mm.KV.ListKeys(pageNumber, ListKeysPerPage)
		if err != nil {
			return errors.Wrapf(err, "failed to list keys - page, %d", pageNumber)
		}
		if len(keys) == 0 {
			return nil
		}

		for _, key := range keys {
			if len(key) != hashKeyLength {
				continue
			}

			allMatch := true
			for _, f := range matchf {
				prefix, appID, userID, namespace, idhash, _ := ParseHashkey(key)
				if !f(prefix, appID, userID, namespace, idhash) {
					allMatch = false
					break
				}
			}
			if len(matchf) > 0 && !allMatch {
				continue
			}

			err = processf(key)
			if err != nil {
				return err
			}
		}
	}
}

func (s *Service) RemoveAllKVAndUserDataForApp(r *incoming.Request, appID apps.AppID) error {
	mm := s.conf.MattermostAPI()
	if err := s.ListHashKeys(r, mm.KV.Delete, WithAppID(appID), WithPrefix(KVAppPrefix)); err != nil {
		return errors.Wrap(err, "failed to remove all data for app")
	}
	if err := s.ListHashKeys(r, mm.KV.Delete, WithAppID(appID), WithPrefix(KVUserPrefix)); err != nil {
		return errors.Wrap(err, "failed to remove all data for app")
	}
	return nil
}

func WithPrefix(prefix string) func(string, apps.AppID, string, string, string) bool {
	return func(p string, _ apps.AppID, _, _, _ string) bool {
		return prefix == "" || p == prefix
	}
}

func WithAppID(appID apps.AppID) func(string, apps.AppID, string, string, string) bool {
	return func(_ string, a apps.AppID, _, _, _ string) bool {
		return appID == "" || a == appID
	}
}

func WithUserID(userID string) func(string, apps.AppID, string, string, string) bool {
	return func(_ string, _ apps.AppID, u, _, _ string) bool {
		return userID == "" || u == userID
	}
}
func WithNamespace(namespace string) func(string, apps.AppID, string, string, string) bool {
	return func(_ string, _ apps.AppID, _, ns, _ string) bool {
		return namespace == "" || ns == namespace
	}
}
