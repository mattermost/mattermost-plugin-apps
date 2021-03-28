package store

import (
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
	return s.mm.KV.Set(s.kvKey(namespace, prefix, id), data)
}

func (s *appKVStore) Get(namespace, prefix, id string, ref interface{}) error {
	return s.mm.KV.Get(s.kvKey(namespace, prefix, id), ref)
}

func (s *appKVStore) Delete(namespace, prefix, id string) error {
	return s.mm.KV.Delete(s.kvKey(namespace, prefix, id))
}

func (s *appKVStore) kvKey(namespace, prefix, id string) string {
	return s.hashkey(config.KVAppPrefix, namespace, prefix, id)
}
