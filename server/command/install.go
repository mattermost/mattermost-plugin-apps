// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"github.com/pkg/errors"
	"github.com/spf13/pflag"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/http/dialog"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

func (s *service) executeInstall(params *params) (*model.CommandResponse, error) {
	appSecret := ""
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&appSecret, "app-secret", "", "App secret")
	if err := fs.Parse(params.current); err != nil {
		return errorOut(params, err)
	}

	if !s.mm.User.HasPermissionTo(params.commandArgs.UserId, model.PERMISSION_MANAGE_SYSTEM) {
		return errorOut(params, utils.ErrForbidden)
	}

	if len(params.current) == 0 {
		return errorOut(params, errors.New("you need to specify the app id"))
	}

	appID := params.current[0]

	m, err := s.proxy.GetManifest(apps.AppID(appID))
	if err != nil {
		return errorOut(params, errors.Wrap(err, "manifest not found"))
	}

	return s.installApp(m, appSecret, params)
}

func (s *service) installApp(m *apps.Manifest, appSecret string, params *params) (*model.CommandResponse, error) {
	conf := s.conf.GetConfig()

	// Finish the installation when the Dialog is submitted, see
	// <plugin>/http/dialog/install.go
	err := s.mm.Frontend.OpenInteractiveDialog(
		dialog.NewInstallAppDialog(m, appSecret, conf.PluginURL, params.commandArgs))
	if err != nil {
		return errorOut(params, errors.Wrap(err, "couldn't open an interactive dialog"))
	}

	return &model.CommandResponse{
		Text:         "please continue by filling out the interactive form",
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, nil
}
