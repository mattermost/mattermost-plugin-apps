// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import pluginapi "github.com/mattermost/mattermost-plugin-api"

type Service struct {
	Configurator Configurator
	Mattermost   *pluginapi.Client
	Proxy        Proxy
	Admin        Admin
	AppServices  AppServices
}
