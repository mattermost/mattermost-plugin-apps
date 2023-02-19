// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"encoding/ascii85"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
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
	// KVPrefix is the global namespace for apps KV data.
	KVPrefix = ".k"

	// UserPrefix is the global namespace used to store user records.
	UserPrefix = ".u"

	// CachedPrefix is the global namespace for storing synchronized cached
	// lists of records like apps and subscriptions.
	CachedPrefix = ".cached"

	// OAuth2StatePrefix is the global namespace used to store OAuth2 ephemeral
	// state data.
	OAuth2StatePrefix = ".o"

	TokenPrefix = ".t"

	DebugPrefix = ".debug."

	// OAuth2UserKey is the key to store OAuth2 user records, in the UserPrefix
	// global namespace. There is only one key per user.
	OAuth2UserKey = "oauth2_user"

	// CallOnceKey and KVClusterMutexKey are used for invoking App Calls once,
	// usually upon a Mattermost instance startup.
	CallOnceKey     = "CallOnce"
	ClusterMutexKey = "Cluster_Mutex"
)

const (
	AppStoreName          = "apps"
	ManifestStoreName     = "manifests"
	SubscriptionStoreName = "subscriptions"
)

const (
	ListKeysPerPage = 1000
)

const (
	hashKeyLength = 82
)

type Service struct {
	App          Apps
	Subscription *SubscriptionStore
	Manifest     *ManifestStore
	AppKV        *KVStore
	OAuth2       *OAuth2Store
	Session      Sessions

	cluster *CachedStoreCluster

	// testReportChan is used in the cluster test to receive test reports from
	// other hosts. It is initialized only on the host that actually runs the
	// test. It receives reports from all nodes in the cluster. It should be
	// synchronized, but is modified only from the test command, rarely, so it's
	// ok not to.
	testReportChan chan CachedStoreTestReport
	testStore      CachedStore[testDataType]
	testDataMutex  *sync.RWMutex
}

// MakeService creates and initializes a persistent storage Service. defaultKind
// will be used to make the app, manifest, and subscription stores and specifies
// how they will be replicated across cluster nodes.
func MakeService(conf config.Service, defaultKind CachedStoreClusterKind) (*Service, error) {
	s := &Service{
		cluster: NewCachedStoreCluster(conf.API(), defaultKind),
		AppKV:   &KVStore{},
		OAuth2:  &OAuth2Store{},
		Session: &SessionStore{},

		testDataMutex: &sync.RWMutex{},
	}
	log := conf.NewBaseLogger()

	var err error
	s.App, err = s.makeAppStore(conf.Get().PluginManifest.Version, log)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize the app store")
	}
	s.Manifest, err = s.makeManifestStore(log)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize the manifest store")
	}
	s.Subscription, err = s.makeSubscriptionStore(log)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize the subscription store")
	}

	return s, nil
}

// Makes a hash key that is used to store data in the KV store, that is specific
// to a given app and user. key is hashed, everything else is preserved in the
// output.
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
		return "", errors.New("appID %s is tooo long, must be 32 bytes or fewer")
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

func hashkey(globalNamespace, appID, userID, appNamespace, key []byte) string {
	hash := make([]byte, 16)
	sha3.ShakeSum128(hash, key)
	encodedID := make([]byte, ascii85.MaxEncodedLen(len(hash)))
	_ = ascii85.Encode(encodedID, hash)

	hashed := make([]byte, 0, model.KeyValueKeyMaxRunes)
	hashed = append(hashed, globalNamespace...)
	hashed = append(hashed, appID...)
	hashed = append(hashed, userID...)
	hashed = append(hashed, appNamespace...)
	hashed = append(hashed, encodedID...)
	return string(hashed)
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

func ListHashKeys(
	r *incoming.Request,
	processf func(key string) error,
	matchf ...func(prefix string, _ apps.AppID, userID, namespace, idhash string) bool,
) error {
	for pageNumber := 0; ; pageNumber++ {
		keys, err := r.API.Mattermost.KV.ListKeys(pageNumber, ListKeysPerPage)
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

func RemoveAllKVAndUserDataForApp(r *incoming.Request, appID apps.AppID) error {
	if err := ListHashKeys(r, r.API.Mattermost.KV.Delete, WithAppID(appID), WithPrefix(KVPrefix)); err != nil {
		return errors.Wrap(err, "failed to remove all data for app")
	}
	if err := ListHashKeys(r, r.API.Mattermost.KV.Delete, WithAppID(appID), WithPrefix(UserPrefix)); err != nil {
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
