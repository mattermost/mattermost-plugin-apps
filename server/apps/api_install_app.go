// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type InInstallApp struct {
	App                App
	Context            CallContext
	GrantedPermissions Permissions
}

func (s *Service) InstallApp(in *InInstallApp) (*App, md.MD, error) {
	// TODO check if acting user is a sysadmin

	app := in.App
	app.GrantedPermissions = app.Manifest.RequestedPermissions

	err := s.Registry.Store(&app)
	if err != nil {
		return nil, "", err
	}

	in.Context.AppID = app.Manifest.AppID
	expApp := app
	expApp.Manifest = nil
	expApp.Secret = ""

	resp, err := s.PostWish(
		Call{
			Wish: app.Manifest.Install,
			Data: &CallData{
				Values:  FormValues{},
				Context: in.Context,
				Expanded: &Expanded{
					App: &expApp,
				},
			},
		})
	if err != nil {
		return nil, "", errors.Wrap(err, "Install failed")
	}

	return &app, resp.Markdown, nil
}
