// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type SessionToken string

type Admin interface {
	ListApps() ([]*App, md.MD, error)
	InstallApp(*Context, SessionToken, *InInstallApp) (*App, md.MD, error)
	ProvisionApp(*Context, SessionToken, *InProvisionApp) (*App, md.MD, error)
}

type InInstallApp struct {
	GrantedPermissions Permissions `json:"granted_permissions,omitempty"`
	GrantedLocations   Locations   `json:"granted_locations,omitempty"`
	AppSecret          string      `json:"app_secret,omitempty"`
	OAuth2TrustedApp   bool        `json:"oauth2_trusted_app,omitempty"`
}

type InProvisionApp struct {
	Manifest  *Manifest `json:"manifest"`
	AppSecret string    `json:"app_secret,omitempty"`
	Force     bool      `json:"force,omitempty"`
}
