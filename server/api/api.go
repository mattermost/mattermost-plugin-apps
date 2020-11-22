// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"errors"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type SessionToken string

type Proxy interface {
	GetManifest(manifestURL string) (*Manifest, error)
	GetBindings(*Context) ([]*Binding, error)
	Call(*Call) (*CallResponse, error)
	Notify(cc *Context, subj Subject) error

	DebugBuiltInApp(AppID, Upstream)
}

type Admin interface {
	InstallApp(*Context, SessionToken, *InInstallApp) (*App, md.MD, error)
	ProvisionApp(*Context, SessionToken, *InProvisionApp) (*App, md.MD, error)
}

var ErrNotABot = errors.New("not a bot")

type AppServices interface {
	Subscribe(*Subscription) error
	Unsubscribe(*Subscription) error
	KVSet(botUserID, prefix, id string, ref interface{}) (bool, error)
	KVGet(botUserID, prefix, id string, ref interface{}) error
	KVDelete(botUserID, prefix, id string) error
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
