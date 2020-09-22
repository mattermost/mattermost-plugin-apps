// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package appmodel

type Manifest struct {
	AppID                AppID
	DisplayName          string
	Description          string
	RootURL              string
	RequestedPermissions Permissions
	CallbackURL          string
	Homepage             string
	InstallCompleteURL   string
}

type AppID string
