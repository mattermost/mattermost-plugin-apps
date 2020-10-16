// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type SessionToken string

type API interface {
	// Call(*Call) (*CallResponse, error)
	InstallApp(*InInstallApp, *Context, SessionToken) (*store.App, md.MD, error)
	ProvisionApp(*InProvisionApp, *Context, SessionToken) (*store.App, md.MD, error)
	Notify(store.Subject, *Context) error
	GetLocations(userID, channelID string) ([]LocationInt, error)
}

type Service struct {
	Configurator configurator.Service
	Mattermost   *pluginapi.Client
	Store        store.Service
	API          API
	Client       Client
}

type service struct {
	Service
}

func NewService(mm *pluginapi.Client, configurator configurator.Service) *Service {
	s := &service{
		Service: Service{
			Store:        store.NewService(mm, configurator),
			Configurator: configurator,
			Mattermost:   mm,
		},
	}
	s.Client = s
	s.API = s

	return &s.Service
}
