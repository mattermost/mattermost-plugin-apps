// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin/dialog"
)

func (s *service) executeInstallMarketplace(params *commandParams) (*model.CommandResponse, error) {
	loc := s.i18n.GetUserLocalizer(params.commandArgs.UserId)
	if len(params.current) == 0 {
		return s.errorOut(params, errors.New(s.i18n.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "apps.command.install.appID",
			Other: "you must specify the app id",
		})))
	}
	appID := apps.AppID(params.current[0])

	m, err := s.proxy.GetManifest(appID)
	if err != nil {
		return s.errorOut(params, errors.Wrap(err, s.i18n.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "apps.command.install.marketplace.error.manifest",
			Other: "manifest not found",
		})))
	}

	return s.installApp(m, "", params)
}

func (s *service) executeInstallAWS(params *commandParams) (*model.CommandResponse, error) {
	loc := s.i18n.GetUserLocalizer(params.commandArgs.UserId)
	if len(params.current) == 0 {
		return s.errorOut(params, errors.New(s.i18n.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "apps.command.install.error.appID",
			Other: "you must specify the app id",
		})))
	}
	appID := apps.AppID(params.current[0])

	if len(params.current) < 2 {
		return s.errorOut(params, errors.New(s.i18n.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "apps.command.install.aws.error.version",
			Other: "you must specify the app version",
		})))
	}
	version := apps.AppVersion(params.current[1])

	m, err := s.proxy.GetManifestFromS3(appID, version)
	if err != nil {
		return s.errorOut(params, errors.Wrap(err, s.i18n.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "apps.command.install.aws.error.manifest",
			Other: "failed to get manifest from S3",
		})))
	}

	_, err = s.proxy.AddLocalManifest(params.commandArgs.UserId, m)
	if err != nil {
		return s.errorOut(params, err)
	}

	return s.installApp(m, "", params)
}

func (s *service) executeInstallHTTP(params *commandParams) (*model.CommandResponse, error) {
	loc := s.i18n.GetUserLocalizer(params.commandArgs.UserId)
	appSecret := ""
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&appSecret, "app-secret", "", "App secret")
	err := fs.Parse(params.current)
	if err != nil {
		return s.errorOut(params, err)
	}

	if len(params.current) == 0 {
		return s.errorOut(params, errors.New(s.i18n.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "apps.command.install.http.error.url",
			Other: "you must specify a manifest URL",
		})))
	}
	manifestURL := params.current[0]

	// Trust the URL only in dev mode
	conf := s.conf.GetConfig()
	data, err := s.httpOut.GetFromURL(manifestURL, conf.DeveloperMode)
	if err != nil {
		return s.errorOut(params, err)
	}

	m, err := apps.ManifestFromJSON(data)
	if err != nil {
		return s.errorOut(params, err)
	}

	_, err = s.proxy.AddLocalManifest(params.commandArgs.UserId, m)
	if err != nil {
		return s.errorOut(params, err)
	}

	return s.installApp(m, appSecret, params)
}

func (s *service) installApp(m *apps.Manifest, appSecret string, params *commandParams) (*model.CommandResponse, error) {
	loc := s.i18n.GetUserLocalizer(params.commandArgs.UserId)
	conf := s.conf.GetConfig()

	// Finish the installation when the Dialog is submitted, see
	// <plugin>/http/dialog/install.go
	err := s.mm.Frontend.OpenInteractiveDialog(
		dialog.NewInstallAppDialog(m, appSecret, conf, params.commandArgs))
	if err != nil {
		return s.errorOut(params, errors.Wrap(err, s.i18n.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "apps.command.install.error.openDialog",
			Other: "couldn't open an interactive dialog",
		})))
	}

	return &model.CommandResponse{
		Text: s.i18n.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "apps.command.install.fillDialog",
			Other: "please continue by filling out the interactive form",
		}),
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, nil
}
