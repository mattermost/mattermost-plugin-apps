// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

const prefixInstalledApp = "app_"
const prefixLocalManifest = "manifest_"
const prefixSubs = "sub_"

type Service struct {
	App          App
	Subscription Subscription
	Manifest     Manifest

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
	s.Subscription = &subscriptionStore{
		Service: s,
	}
	s.Manifest = &manifestStore{
		Service: s,
	}
	return s
}
