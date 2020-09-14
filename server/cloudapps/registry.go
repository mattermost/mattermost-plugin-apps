// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package cloudapps

import (
	"github.com/mattermost/mattermost-plugin-cloudapps/server/configurator"
)

type Registry interface {
	InstallApp(*InInstallApp) (*OutInstallApp, error)
	ListApps() (*OutListApps, error)
	GetApp(AppID) (*App, error)
}

type registry struct {
	configurator.Configurator

	// <><> Needs to come from config to be synchronized, or read from KV every request, sync.Map is unnecessary
	apps map[AppID]*App
}

var _ Registry = (*registry)(nil)

func NewRegistry(configurator configurator.Configurator) Registry {
	return &registry{
		Configurator: configurator,
	}
}
