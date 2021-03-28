package store

import (
	"crypto/md5" // nolint:gosec
	"encoding/base64"
	"path"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

type AppKVStore interface {
	Set(namespace, prefix, id string, data []byte) (bool, error)
	Get(namespace, prefix, id string, ref interface{}) error
	Delete(namespace, prefix, id string) error
}

type appKVStore struct {
	*Service
}

var _ AppKVStore = (*appKVStore)(nil)

// TODO use raw byte API: for now all JSON is re-encoded to use api.Mattermost API
func (s *appKVStore) Set(namespace, prefix, id string, data []byte) (bool, error) {
	return s.mm.KV.Set(kvKey(namespace, prefix, id), data)
}

func (s *appKVStore) Get(namespace, prefix, id string, ref interface{}) error {
	return s.mm.KV.Get(kvKey(namespace, prefix, id), ref)
}

func (s *appKVStore) Delete(namespace, prefix, id string) error {
	return s.mm.KV.Delete(kvKey(namespace, prefix, id))
}

func kvKey(namespace, prefix, id string) string {
	if id == "" || namespace == "" {
		return ""
	}

	namespacePrefixHash := md5.Sum([]byte(namespace + prefix)) // nolint:gosec
	idHash := md5.Sum([]byte(id))                              // nolint:gosec
	key := config.KVAppPrefix + path.Join(
		base64.RawURLEncoding.EncodeToString(namespacePrefixHash[:]),
		base64.RawURLEncoding.EncodeToString(idHash[:]))

	if len(key) > model.KEY_VALUE_KEY_MAX_RUNES {
		return key[:model.KEY_VALUE_KEY_MAX_RUNES]
	}

	return key
}
