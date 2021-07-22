package store

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

const (
	keysPerPage = 1000
)

type AppKVStore interface {
	Set(botUserID, prefix, id string, ref interface{}) (bool, error)
	Get(botUserID, prefix, id string, ref interface{}) error
	Delete(botUserID, prefix, id string) error
	DeleteAll(botUserID string) error
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

func (s *appKVStore) DeleteAll(botUserID string) error {
	prefix := config.KVAppPrefix + botUserID
	var keysToDelete []string

	for i := 0; ; i++ {
		keys, err := s.mm.KV.ListKeys(i, keysPerPage)
		if err != nil {
			return errors.Wrapf(err, "failed to list keys for deletion - page, %d", i)
		}

		for _, k := range keys {
			if strings.HasPrefix(k, prefix) {
				keysToDelete = append(keysToDelete, k)
			}
		}

		if len(keys) < keysPerPage {
			break
		}
	}

	for _, k := range keysToDelete {
		err := s.mm.KV.Delete(k)
		if err != nil {
			return errors.Wrap(err, "failed to delete key")
		}
	}

	return nil
}
