package store

import (
	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

type AppKVStore interface {
	Set(botUserID, prefix, id string, ref interface{}) (bool, error)
	Get(botUserID, prefix, id string, ref interface{}) error
	Delete(botUserID, prefix, id string) error
}

type appKVStore struct {
	*Service
}

var _ AppKVStore = (*appKVStore)(nil)

// TODO use raw byte API: for now all JSON is re-encoded to use api.Mattermost API
func (s *appKVStore) Set(botUserID, prefix, id string, ref interface{}) (bool, error) {
	key, err := s.hashkey(config.KVAppPrefix, botUserID, prefix, id)
	if err != nil {
		return false, err
	}
	return s.mm.KV.Set(key, ref)
}

func (s *appKVStore) Get(botUserID, prefix, id string, ref interface{}) error {
	key, err := s.hashkey(config.KVAppPrefix, botUserID, prefix, id)
	if err != nil {
		return err
	}
	return s.mm.KV.Get(key, ref)
}

func (s *appKVStore) Delete(botUserID, prefix, id string) error {
	key, err := s.hashkey(config.KVAppPrefix, botUserID, prefix, id)
	if err != nil {
		return err
	}
	return s.mm.KV.Delete(key)
}
