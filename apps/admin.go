// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

type SessionToken string

type InInstallApp struct {
	AppID            AppID  `json:"app_id"`
	AppSecret        string `json:"app_secret,omitempty"`
	OAuth2TrustedApp bool   `json:"oauth2_trusted_app,omitempty"`
}
