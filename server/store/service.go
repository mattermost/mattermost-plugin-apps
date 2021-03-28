// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"

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
