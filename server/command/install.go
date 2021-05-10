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

func (s *service) executeInstall(params *params) (*model.CommandResponse, error) {
	appSecret := ""
	manifestURL := ""
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&appSecret, "app-secret", "", "App secret")
	fs.StringVar(&manifestURL, "url", "", "App's manifest URL")
	err := fs.Parse(params.current)
	if err != nil {
		return errorOut(params, err)
	}

	var m *apps.Manifest
	var appID apps.AppID
	conf := s.conf.GetConfig()
	if conf.MattermostCloudMode {
		if len(params.current) == 0 {
			return errorOut(params, errors.New("you must specify the app id"))
		}
		appID = apps.AppID(params.current[0])
	} else {
		if manifestURL == "" {
			return errorOut(params, errors.New("you must add a `--url`"))
		}
		// Trust the URL only in dev mode
		var data []byte
		data, err = s.httpOut.GetFromURL(manifestURL, conf.DeveloperMode)
		if err != nil {
			return errorOut(params, err)
		}
		m, err = apps.ManifestFromJSON(data)
		if err != nil {
			return errorOut(params, err)
		}

		_, err = s.proxy.AddLocalManifest(params.commandArgs.UserId, m)
		if err != nil {
			return errorOut(params, err)
		}
		appID = m.AppID
	}

	// Get the manifest from the store, even if redundant
	m, err = s.proxy.GetManifest(appID)
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
