// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

var ErrNotABot = errors.New("not a bot")

type Service struct {
	Mattermost   *pluginapi.Client
	API          API
	Configurator Configurator
	Upstream     Upstream
}
type SessionToken string

type API interface {
	Call(*Call) (*CallResponse, error)
	GetBindings(*Context) ([]*Binding, error)
	InstallApp(*Context, SessionToken, *InInstallApp) (*App, md.MD, error)
	Notify(cc *Context, subj Subject) error
	ProvisionApp(*Context, SessionToken, *InProvisionApp) (*App, md.MD, error)
	Subscribe(*Subscription) error
	Unsubscribe(*Subscription) error

	ListApps() []*App
	GetApp(appID AppID) (*App, error)
	StoreApp(app *App) error

	KVGet(namespace, prefix, id string, ref interface{}) error
	KVSet(namespace, prefix, id string, ref interface{}) (bool, error)
	KVDelete(namespace, prefix, id string) error
}

type InInstallApp struct {
	GrantedPermissions Permissions `json:"granted_permissions,omitempty"`
	GrantedLocations   Locations   `json:"granted_locations,omitempty"`
	AppSecret          string      `json:"app_secret,omitempty"`
	OAuth2TrustedApp   bool        `json:"oauth2_trusted_app,omitempty"`
}

type InProvisionApp struct {
	ManifestURL string `json:"manifest_url,omitempty"`
	AppSecret   string `json:"app_secret,omitempty"`
	Force       bool   `json:"force,omitempty"`
}
