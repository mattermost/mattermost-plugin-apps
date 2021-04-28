// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"encoding/ascii85"
	"fmt"
	"path"

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

func NewService(mm *pluginapi.Client, conf config.Service) *Service {
	s := &Service{
		mm:   mm,
		conf: conf,
	}
	s.App = &appStore{
		Service: s,
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
	s.Manifest = &manifestStore{
		Service: s,
	}
	return s
}

func (s *Service) hashkey(globalPrefix, namespace, prefix, id string) string {
	if id == "" || namespace == "" {
		return ""
	}
	key := hashkey(globalPrefix, namespace, prefix, id)
	if len(key) > model.KEY_VALUE_KEY_MAX_RUNES {
		s.mm.Log.Info(fmt.Sprintf("AppKV key truncated by %v characters", len(key)-model.KEY_VALUE_KEY_MAX_RUNES),
			"namespace", namespace,
			"prefix", prefix,
			"id", id)
		return key[:model.KEY_VALUE_KEY_MAX_RUNES]
	}

	return key
}

func hashkey(globalPrefix, namespace, prefix, id string) string {
	namespacePrefixHash := make([]byte, 20)
	sha3.ShakeSum128(namespacePrefixHash, []byte(namespace+prefix))

	idHash := make([]byte, 16)
	sha3.ShakeSum128(idHash, []byte(id))

	encodedPrefix := make([]byte, ascii85.MaxEncodedLen(len(namespacePrefixHash)))
	_ = ascii85.Encode(encodedPrefix, namespacePrefixHash)

	encodedID := make([]byte, ascii85.MaxEncodedLen(len(idHash)))
	_ = ascii85.Encode(encodedID, idHash)

	return globalPrefix + path.Join(string(encodedPrefix), string(encodedID))
}
