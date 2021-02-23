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

// Manifest is loaded from a URL for convenience, it really should be provided
// as text/JSON or as a file.
func (s *service) executeInstallHTTPApp(params *params) (*model.CommandResponse, error) {
	manifestURL := ""
	appSecret := ""
	force := false
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&manifestURL, "url", "", "manifest URL")
	fs.StringVar(&appSecret, "app-secret", "", "App secret")
	fs.BoolVar(&force, "force", false, "Force re-installing of the app")

	err := fs.Parse(params.current)
	if err != nil {
		return errorOut(params, err)
	}

	if !s.api.Mattermost.User.HasPermissionTo(params.commandArgs.UserId, model.PERMISSION_MANAGE_SYSTEM) {
		return errorOut(params, errors.New("forbidden"))
	}

	manifest, err := proxy.LoadManifest(manifestURL)
	if err != nil {
		return errorOut(params, err)
	}
	return s.installApp(manifest, appSecret, force, params)
}

func (s *service) executeInstallAWSApp(params *params) (*model.CommandResponse, error) {
	appID := ""
	appSecret := ""
	force := false

	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&appID, "app-id", "", "ID of the app")
	fs.StringVar(&appSecret, "app-secret", "", "App secret")
	fs.BoolVar(&force, "force", false, "Force re-provisioning of the app")

	err := fs.Parse(params.current)
	if err != nil {
		return errorOut(params, err)
	}

	if !s.api.Mattermost.User.HasPermissionTo(params.commandArgs.UserId, model.PERMISSION_MANAGE_SYSTEM) {
		return errorOut(params, errors.New("forbidden"))
	}

	app, err := s.api.Admin.GetApp(apps.AppID(appID))
	if err != nil {
		return errorOut(params, errors.Wrap(err, "App not found"))
	}
	if app.Manifest == nil || app.Manifest.Type != apps.AppTypeAWSLambda {
		return errorOut(params, errors.Wrap(err, "Not an AWS app"))
	}
	return s.installApp(app.Manifest, appSecret, force, params)
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

func (s *service) executeProvisionAWSApp(params *params) (*model.CommandResponse, error) {
	releaseURL := ""
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&releaseURL, "url", "", "release URL")

	err := fs.Parse(params.current)
	if err != nil {
		return errorOut(params, err)
	}

	err = s.api.AWS.ProvisionApp(releaseURL)
	if err != nil {
		return errorOut(params, err)
	}

	return &model.CommandResponse{
		Text:         fmt.Sprintf("installed lambda functions from url %s.", releaseURL),
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, nil
}
