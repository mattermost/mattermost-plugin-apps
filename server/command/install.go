// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"

	"github.com/mattermost/mattermost-server/v5/model"

	sdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/aws"
	"github.com/mattermost/mattermost-plugin-apps/server/http/dialog"
)

func (s *service) executeInstall(params *params) (*model.CommandResponse, error) {
	manifestURL := ""
	appSecret := ""
	force := false
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&manifestURL, "url", "", "manifest URL")
	fs.StringVar(&appSecret, "app-secret", "", "App secret")
	fs.BoolVar(&force, "force", false, "Force re-provisioning of the app")

	err := fs.Parse(params.current)
	if err != nil {
		return normalOut(params, nil, err)
	}

	manifest, err := s.api.Proxy.GetManifest(manifestURL)
	if err != nil {
		return normalOut(params, nil, err)
	}

	app, _, err := s.api.Admin.ProvisionApp(
		&api.Context{
			ActingUserID: params.commandArgs.UserId,
		},
		api.SessionToken(params.commandArgs.Session.Token),
		&api.InProvisionApp{
			Manifest:  manifest,
			AppSecret: appSecret,
			Force:     force,
		},
	)
	if err != nil {
		return normalOut(params, nil, err)
	}

	conf := s.api.Configurator.GetConfig()

	// Finish the installation when the Dialog is submitted, see
	// <plugin>/http/dialog/install.go
	err = s.api.Mattermost.Frontend.OpenInteractiveDialog(
		dialog.NewInstallAppDialog(manifest, appSecret, conf.PluginURL, params.commandArgs))
	if err != nil {
		return normalOut(params, nil, errors.Wrap(err, "couldn't open an interactive dialog"))
	}

	team, err := s.api.Mattermost.Team.Get(params.commandArgs.TeamId)
	if err != nil {
		return normalOut(params, nil, err)
	}

	return &model.CommandResponse{
		GotoLocation: params.commandArgs.SiteURL + "/" + team.Name + "/messages/@" + app.BotUsername,
		Text:         fmt.Sprintf("redirected to the DM with @%s to continue installing **%s**", app.BotUsername, manifest.DisplayName),
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, nil
}

func (s *service) executeExperimentalInstall(params *params) (*model.CommandResponse, error) {
	releaseURL := ""
	secret := ""
	id := ""
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&releaseURL, "url", "", "release URL")
	fs.StringVar(&id, "id", "", "AWS Access Key ID")
	fs.StringVar(&secret, "secret", "", "AWS Secret access key")

	err := fs.Parse(params.current)
	if err != nil {
		return normalOut(params, nil, err)
	}

	config := &sdk.Config{
		Region:      sdk.String("us-east-2"),
		Credentials: credentials.NewStaticCredentials(id, secret, ""),
	}
	logger := log.New()
	logger.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	// Output to stdout instead of the default stderr.
	logger.SetOutput(os.Stdout)

	client := aws.NewAWSClientWithConfig(config, logger)
	if err = client.InstallApp(releaseURL); err != nil {
		return normalOut(params, nil, err)
	}
	// An example of the function Invoke
	// res, err := client.InvokeFunction("my_app_my_func", "blabla")
	// if err != nil {
	// 	return normalOut(params, nil, err)
	// }

	return &model.CommandResponse{
		Text:         fmt.Sprintf("installed lambda functions from url %s.", releaseURL),
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, nil
}
