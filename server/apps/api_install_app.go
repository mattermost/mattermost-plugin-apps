// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
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

func (s *Service) InstallApp(in *InInstallApp) (*OutInstallApp, error) {
	if in.Manifest.AppID == "" {
		return nil, errors.New("app ID must not be empty")
	}
	// TODO check if acting user is a sysadmin

	// TODO remove mock, implement for real
	app := &App{
		Manifest:               in.Manifest,
		GrantedPermissions:     in.Manifest.RequestedPermissions,
		NoUserConsentForOAuth2: in.NoUserConsentForOAuth2,
		Secret:                 in.Secret,
	}

	err := s.Registry.Store(app)
	if err != nil {
		return nil, err
	}

	// TODO expand CallData
	resp, err := s.PostWish(app.Manifest.AppID, in.ActingMattermostUserID, app.Manifest.Install, &CallData{})
	if err != nil {
		return nil, errors.Wrap(err, "Install failed")
	}

	out := &OutInstallApp{
		MD:  resp.Markdown,
		App: app,
	}
	return out, nil
}
