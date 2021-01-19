// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package modelapps

type SessionToken string

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
