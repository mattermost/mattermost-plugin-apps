// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package appservices

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

type AppServices struct {
	mm    *pluginapi.Client
	conf  api.Configurator
	store api.Store
}

var _ api.AppServices = (*AppServices)(nil)

func NewAppServices(mm *pluginapi.Client, conf api.Configurator, store api.Store) *AppServices {
	return &AppServices{mm, conf, store}
}
