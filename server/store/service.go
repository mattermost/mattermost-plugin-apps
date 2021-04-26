// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"crypto/md5" // nolint:gosec
	"encoding/base64"
	"fmt"
	"path"

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

	namespacePrefixHash := md5.Sum([]byte(namespace + prefix)) // nolint:gosec
	idHash := md5.Sum([]byte(id))                              // nolint:gosec
	key := globalPrefix + path.Join(
		base64.RawURLEncoding.EncodeToString(namespacePrefixHash[:]),
		base64.RawURLEncoding.EncodeToString(idHash[:]))

	if len(key) > model.KEY_VALUE_KEY_MAX_RUNES {
		s.mm.Log.Info(fmt.Sprintf("AppKV key truncated by %v characters", len(key)-model.KEY_VALUE_KEY_MAX_RUNES),
			"namespace", namespace,
			"prefix", prefix,
			"id", id)
		return key[:model.KEY_VALUE_KEY_MAX_RUNES]
	}

	return key
}
