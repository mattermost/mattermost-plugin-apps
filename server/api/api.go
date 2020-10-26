// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type SessionToken string

type API interface {
	GetLocations(userID, channelID string) ([]LocationInt, error)

	ProvisionApp(*InProvisionApp, *Context, SessionToken) (*App, md.MD, error)
	InstallApp(*InInstallApp, *Context, SessionToken) (*App, md.MD, error)

	Call(*Call) (*CallResponse, error)
	NotifySubscribedApps(subj Subject, cc *Context) error
}

type InInstallApp struct {
	GrantedPermissions Permissions
	AppSecret          string
	OAuth2TrustedApp   bool
}

type InProvisionApp struct {
	ManifestURL string
	AppSecret   string
	Force       bool
}
