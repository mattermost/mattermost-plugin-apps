package store

import (
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type AppKVStore interface {
	Set(_ *incoming.Request, prefix, id string, data []byte) (bool, error)
	Get(_ *incoming.Request, prefix, id string) ([]byte, error)
	Delete(_ *incoming.Request, prefix, id string) error
	List(_ *incoming.Request, namespace string, processf func(key string) error) error
}

type appKVStore struct {
	*Service
}

var _ AppKVStore = (*appKVStore)(nil)

func (s *appKVStore) Set(r *incoming.Request, prefix, id string, data []byte) (bool, error) {
	if r.SourceAppID() == "" || r.ActingUserID() == "" {
		return false, utils.NewInvalidError("source app ID or user ID missing in the request")
	}
	key, err := Hashkey(KVAppPrefix, r.SourceAppID(), r.ActingUserID(), prefix, id)
	if err != nil {
		return false, err
	}

	set, err := s.conf.MattermostAPI().KV.Set(key, data)
	if err != nil {
		return false, err
	}
	if set {
		r.Log.Debugw("AppKV set", "prefix", prefix, "id", id, "hashkey", key)
	}
	return set, nil
}

func (s *appKVStore) Get(r *incoming.Request, prefix, id string) ([]byte, error) {
	key, err := Hashkey(KVAppPrefix, r.SourceAppID(), r.ActingUserID(), prefix, id)
	if err != nil {
		return nil, err
	}

	var data []byte
	if err = s.conf.MattermostAPI().KV.Get(key, &data); err != nil {
		return nil, err
	}

	return data, err
}

func (s *appKVStore) Delete(r *incoming.Request, prefix, id string) error {
	key, err := Hashkey(KVAppPrefix, r.SourceAppID(), r.ActingUserID(), prefix, id)
	if err != nil {
		return err
	}

	err = s.conf.MattermostAPI().KV.Delete(key)
	if err != nil {
		return err
	}
	r.Log.Debugw("AppKV deleted", "prefix", prefix, "id", id, "hashkey", key)
	return nil
}

func (s *appKVStore) List(r *incoming.Request, namespace string, processf func(key string) error) error {
	return s.ListHashKeys(r, processf,
		WithPrefix(KVAppPrefix),
		WithAppID(r.SourceAppID()),
		WithUserID(r.ActingUserID()),
		WithNamespace(namespace))
}
