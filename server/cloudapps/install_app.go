// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package cloudapps

import (
	"errors"

	"github.com/mattermost/mattermost-plugin-cloudapps/server/utils/md"
)

type InInstallApp struct {
	ActingMattermostUserID string
	NoUserConsentForOAuth2 bool
	Manifest               *Manifest
	Secret                 string
}

type OutInstallApp struct {
	md.MD
	App *App
}

func (r *registry) InstallApp(in *InInstallApp) (*OutInstallApp, error) {
	if in.Manifest.AppID == "" {
		return nil, errors.New("app ID must not be empty")
	}
	// TODO check if acting user is a sysadmin

	// TODO remove mock, implement for real
	app := &App{
		Manifest:               in.Manifest,
		GrantedPermissions:     in.Manifest.RequestedPermissions,
		NoUserConsentForOAuth2: in.NoUserConsentForOAuth2,
	}
	r.apps[in.Manifest.AppID] = app

	out := &OutInstallApp{
		MD:  md.Markdownf("Installed %s (%s)", in.Manifest.DisplayName, in.Manifest.AppID),
		App: app,
	}
	return out, nil
}
