// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type SessionToken string

type API interface {
	GetBindings(*Context) ([]*Binding, error)

	ProvisionApp(*Context, SessionToken, *InProvisionApp) (*App, md.MD, error)
	InstallApp(*Context, SessionToken, *InInstallApp) (*App, md.MD, error)

	Call(*Call) (*CallResponse, error)
	Notify(cc *Context, subj Subject) error
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
