// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type InInstallApp struct {
	ActingMattermostUserID string
	App                    *App
	LogChannelID           string
	LogRootPostID          string
	SessionToken           string
}

type OutInstallApp struct {
	md.MD
	App *App
}

func (s *Service) InstallApp(in *InInstallApp) (*OutInstallApp, error) {
	if in.App.Manifest.AppID == "" {
		return nil, errors.New("app ID must not be empty")
	}
	// TODO check if acting user is a sysadmin

	// TODO remove mock, implement for real
	app := *in.App
	app.GrantedPermissions = app.Manifest.RequestedPermissions

	err := s.Registry.Store(&app)
	if err != nil {
		return nil, err
	}

	// TODO expand CallData
	callData := &CallData{
		Values: FormValues{
			Parsed: map[string]interface{}{
				"X": "Y",
			},
		},
		Env: map[string]interface{}{
			"log_root_post_id": in.LogRootPostID,
			"log_channel_id":   in.LogChannelID,
		},
	}

	resp, err := s.PostWish(app.Manifest.AppID, in.ActingMattermostUserID, app.Manifest.Install, callData)
	if err != nil {
		return nil, errors.Wrap(err, "Install failed")
	}

	out := &OutInstallApp{
		MD:  resp.Markdown,
		App: &app,
	}
	return out, nil
}
