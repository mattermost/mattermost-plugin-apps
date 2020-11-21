// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
)

const prefixSubs = "sub_"

type store struct {
	mm   *pluginapi.Client
	conf apps.Configurator
}

func NewStore(mm *pluginapi.Client, conf apps.Configurator) apps.Store {
	return &store{
		mm:   mm,
		conf: conf,
	}
}
