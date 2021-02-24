// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/http/dialog"
)

func (s *service) executeInstall(params *params) (*model.CommandResponse, error) {
	appID := ""
	manifestURL := ""
	appSecret := ""
	force := false
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&appID, "app-id", "", "ID of the app")
	fs.StringVar(&manifestURL, "url", "", "manifest URL")
	fs.StringVar(&appSecret, "app-secret", "", "App secret")
	fs.BoolVar(&force, "force", false, "Force re-installing of the app")

	if err := fs.Parse(params.current); err != nil {
		return errorOut(params, err)
	}

	if !s.api.Mattermost.User.HasPermissionTo(params.commandArgs.UserId, model.PERMISSION_MANAGE_SYSTEM) {
		return errorOut(params, errors.New("forbidden"))
	}

	if appID != "" {
		app, err := s.api.Admin.GetApp(apps.AppID(appID))
		if err != nil {
			return errorOut(params, errors.Wrap(err, "App not found"))
		}
		if app.Manifest == nil || app.Manifest.Type != apps.AppTypeAWSLambda {
			return errorOut(params, errors.Wrap(err, "Not an AWS app"))
		}
		return s.installApp(app.Manifest, appSecret, force, params)
	}
	manifest, err := proxy.LoadManifest(manifestURL)
	if err != nil {
		return errorOut(params, err)
	}
	if manifest == nil || manifest.Type != apps.AppTypeHTTP {
		return errorOut(params, errors.Wrap(err, "Not an HTTP app"))
	}
	return s.installApp(manifest, appSecret, force, params)
}

func (s *service) installApp(manifest *apps.Manifest, appSecret string, force bool, params *params) (*model.CommandResponse, error) {
	app, _, err := s.api.Admin.ProvisionApp(
		&apps.Context{
			ActingUserID: params.commandArgs.UserId,
		},
		apps.SessionToken(params.commandArgs.Session.Token),
		&apps.InProvisionApp{
			Manifest:  manifest,
			AppSecret: appSecret,
			Force:     force,
		},
	)
	if err != nil {
		return errorOut(params, err)
	}

	conf := s.api.Configurator.GetConfig()

	// Finish the installation when the Dialog is submitted, see
	// <plugin>/http/dialog/install.go
	err = s.api.Mattermost.Frontend.OpenInteractiveDialog(
		dialog.NewInstallAppDialog(app.Manifest, app.Secret, conf.PluginURL, params.commandArgs))
	if err != nil {
		return errorOut(params, errors.Wrap(err, "couldn't open an interactive dialog"))
	}

	team, err := s.api.Mattermost.Team.Get(params.commandArgs.TeamId)
	if err != nil {
		return errorOut(params, err)
	}

	return &model.CommandResponse{
		GotoLocation: params.commandArgs.SiteURL + "/" + team.Name + "/messages/@" + app.BotUsername,
		Text:         fmt.Sprintf("redirected to the DM with @%s to continue installing **%s**", app.BotUsername, app.Manifest.DisplayName),
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, nil
}
