// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/dgrijalva/jwt-go"
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-apps/server/aws"
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

const OutgoingAuthHeader = "Mattermost-App-Authorization"

type Service struct {
	Configurator configurator.Service
	Mattermost   *pluginapi.Client
	API          API
	Client       Client
	AWSProxy     aws.Proxy
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
}

type Client interface {
	GetBindings(*Context) ([]*Binding, error)
	GetManifest(manifestURL string) (*Manifest, error)
	PostCall(*Call) (*CallResponse, error)
	PostNotification(*Notification) error
}

type JWTClaims struct {
	jwt.StandardClaims
	ActingUserID string `json:"acting_user_id,omitempty"`
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
