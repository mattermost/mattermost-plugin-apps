package impl

import (
	"crypto/md5" // nolint:gosec

	"encoding/hex"
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
)

// TODO use raw byte API: for now all JSON is re-encoded to use apps.Mattermost API
func (s *service) KVSet(botUserID, prefix, id string, ref interface{}) (bool, error) {
	changed := false
	var err error
	err = s.kvDo(botUserID, prefix, id, ref, func(key string, ref interface{}) error {
		changed, err = s.Mattermost.KV.Set(key, ref)
		return err
	})
	if err != nil {
		return false, err
	}
	return changed, nil
}

func (s *service) KVGet(botUserID, prefix, id string, ref interface{}) error {
	return s.kvDo(botUserID, prefix, id, ref, s.Mattermost.KV.Get)
}

func (s *service) KVDelete(botUserID, prefix, id string) error {
	return s.kvDo(botUserID, prefix, id, nil, func(key string, _ interface{}) error {
		return s.Mattermost.KV.Delete(key)
	})
}

func kvKey(namespace, prefix, id string) string {
	if id == "" || namespace == "" {
		return ""
	}

	namespaceHash := md5.Sum([]byte(namespace)) // nolint:gosec
	idHash := md5.Sum([]byte(id))               // nolint:gosec
	return strings.Join([]string{
		hex.EncodeToString(namespaceHash[:]),
		prefix,
		hex.EncodeToString(idHash[:]),
	}, "/")
}

// TODO use raw byte API: for now all JSON is re-encoded to use apps.Mattermost API
func (s *service) kvDo(botUserID, prefix, id string, ref interface{}, f func(key string, ref interface{}) error) error {
	mmuser, err := s.Mattermost.User.Get(botUserID)
	if err != nil {
		return err
	}

	if !mmuser.IsBot {
		return apps.ErrNotABot
	}

	return f(kvKey(botUserID, prefix, id), ref)
}
