// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"github.com/pkg/errors"
	"github.com/spf13/pflag"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin/dialog"
)

func (s *service) executeInstallMarketplace(params *commandParams) (*model.CommandResponse, error) {
	if len(params.current) == 0 {
		return errorOut(params, errors.New("you must specify the app id"))
	}
	appID := apps.AppID(params.current[0])

	m, err := s.proxy.GetManifest(appID)
	if err != nil {
		return errorOut(params, errors.Wrap(err, "manifest not found"))
	}

	return s.installApp(m, "", params)
}

func (s *service) executeInstallAWS(params *commandParams) (*model.CommandResponse, error) {
	if len(params.current) == 0 {
		return errorOut(params, errors.New("you must specify the app id"))
	}
	appID := apps.AppID(params.current[0])

	if len(params.current) < 2 {
		return errorOut(params, errors.New("you must specify the app version"))
	}
	version := apps.AppVersion(params.current[1])

	m, err := s.proxy.GetManifestFromS3(appID, version)
	if err != nil {
		return errorOut(params, errors.Wrap(err, "failed to get manifest from S3"))
	}

	_, err = s.proxy.AddLocalManifest(params.commandArgs.UserId, m)
	if err != nil {
		return errorOut(params, err)
	}

	return s.installApp(m, "", params)
}

func (s *service) executeInstallHTTP(params *commandParams) (*model.CommandResponse, error) {
	appSecret := ""
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&appSecret, "app-secret", "", "App secret")
	err := fs.Parse(params.current)
	if err != nil {
		return errorOut(params, err)
	}

	if len(params.current) == 0 {
		return errorOut(params, errors.New("you must specify a manifest URL"))
	}
	manifestURL := params.current[0]

	// Trust the URL only in dev mode
	conf := s.conf.GetConfig()
	data, err := s.httpOut.GetFromURL(manifestURL, conf.DeveloperMode)
	if err != nil {
		return errorOut(params, err)
	}

	m, err := apps.ManifestFromJSON(data)
	if err != nil {
		return errorOut(params, errors.Wrap(err, "unable to decode "+manifestURL))
	}

	_, err = s.proxy.AddLocalManifest(params.commandArgs.UserId, m)
	if err != nil {
		return errorOut(params, err)
	}

	return s.installApp(m, appSecret, params)
}

func (s *service) installApp(m *apps.Manifest, appSecret string, params *commandParams) (*model.CommandResponse, error) {
	conf := s.conf.GetConfig()

	// Finish the installation when the Dialog is submitted, see
	// <plugin>/http/dialog/install.go
	err := s.mm.Frontend.OpenInteractiveDialog(
		dialog.NewInstallAppDialog(m, appSecret, conf, params.commandArgs))
	if err != nil {
		return errorOut(params, errors.Wrap(err, "couldn't open an interactive dialog"))
	}

	return &model.CommandResponse{
		Text:         "please continue by filling out the interactive form",
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, nil
}
