// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package admin

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

type Admin struct {
	mm    *pluginapi.Client
	conf  api.Configurator
	store api.Store
	proxy api.Proxy
}

var _ api.Admin = (*Admin)(nil)

func NewAdmin(mm *pluginapi.Client, conf api.Configurator, store api.Store, proxy api.Proxy) *Admin {
	return &Admin{mm, conf, store, proxy}
}
