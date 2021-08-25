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
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (s *service) executeInstallMarketplace(params *commandParams) *model.CommandResponse {
	if len(params.current) == 0 {
		return s.errorOut(params, utils.NewLocError(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.install.error.appID",
				Other: "you must specify the app id",
			},
		}), nil)
	}
	appID := apps.AppID(params.current[0])

	m, err := s.proxy.GetManifest(appID)
	if err != nil {
		return s.errorOut(params, utils.NewLocError(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.install.marketplace.error.manifest",
				Other: "manifest not found",
			},
		}), errors.Wrap(err, "manifest not found"))
	}

	return s.installApp(m, "", params)
}

func (s *service) executeInstallAWS(params *commandParams) *model.CommandResponse {
	if len(params.current) == 0 {
		return s.errorOut(params, utils.NewLocError(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.install.error.appID",
				Other: "you must specify the app id",
			},
		}), nil)
	}
	appID := apps.AppID(params.current[0])

	if len(params.current) < 2 {
		return s.errorOut(params, utils.NewLocError(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.install.aws.error.version",
				Other: "you must specify the app version",
			},
		}), nil)
	}
	version := apps.AppVersion(params.current[1])

	m, err := s.proxy.GetManifestFromS3(appID, version)
	if err != nil {
		return s.errorOut(params, utils.NewLocError(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.install.aws.error.manifest",
				Other: "failed to get manifest from S3",
			},
		}), errors.Wrap(err, "failed to get manifest from S3"))
	}

	_, err = s.proxy.AddLocalManifest(params.commandArgs.UserId, m)
	if err != nil {
		return s.errorOut(params, nil, err)
	}

	return s.installApp(m, "", params)
}

func (s *service) executeInstallHTTP(params *commandParams) *model.CommandResponse {
	appSecret := ""
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&appSecret, "app-secret", "", "App secret")
	err := fs.Parse(params.current)
	if err != nil {
		return s.errorOut(params, nil, err)
	}

	if len(params.current) == 0 {
		return s.errorOut(params, utils.NewLocError(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.install.http.error.url",
				Other: "you must specify a manifest URL",
			},
		}), nil)
	}
	manifestURL := params.current[0]

	// Trust the URL only in dev mode
	conf := s.conf.Get()
	data, err := s.httpOut.GetFromURL(manifestURL, conf.DeveloperMode)
	if err != nil {
		return s.errorOut(params, nil, err)
	}

	m, err := apps.ManifestFromJSON(data)
	if err != nil {
		return s.errorOut(params, utils.NewLocError(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.install.http.error.decodeManifest",
				Other: "unable to decode {{.ManifestURL}}",
			},
			TemplateData: map[string]interface{}{
				"ManifestURL": manifestURL,
			},
		}), errors.Wrap(err, "unable to decode "+manifestURL))
	}

	_, err = s.proxy.AddLocalManifest(params.commandArgs.UserId, m)
	if err != nil {
		return s.errorOut(params, nil, err)
	}

	return s.installApp(m, appSecret, params)
}

func (s *service) executeInstallKubeless(params *commandParams) *model.CommandResponse {
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	err := fs.Parse(params.current)
	if err != nil {
		return s.errorOut(params, nil, err)
	}
	if len(params.current) == 0 {
		return s.errorOut(params, utils.NewLocError(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.install.http.error.url",
				Other: "you must specify a manifest URL",
			},
		}), nil)
	}
	manifestURL := params.current[0]

	// Trust the URL only in dev mode
	conf := s.conf.Get()
	data, err := s.httpOut.GetFromURL(manifestURL, conf.DeveloperMode)
	if err != nil {
		return s.errorOut(params, nil, err)
	}

	m, err := apps.ManifestFromJSON(data)
	if err != nil {
		return s.errorOut(params, nil, err)
	}

	_, err = s.proxy.AddLocalManifest(params.commandArgs.UserId, m)
	if err != nil {
		return s.errorOut(params, nil, err)
	}

	return s.installApp(m, "", params)
}

func (s *service) installApp(m *apps.Manifest, appSecret string, params *commandParams) *model.CommandResponse {
	conf := s.conf.Get()

	// Finish the installation when the Dialog is submitted, see
	// <plugin>/http/dialog/install.go
	err := s.conf.MattermostAPI().Frontend.OpenInteractiveDialog(
		dialog.NewInstallAppDialog(m, appSecret, conf, params.commandArgs))
	if err != nil {
		return s.errorOut(params, utils.NewLocError(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.install.error.openDialog",
				Other: "couldn't open an interactive dialog",
			},
		}), errors.Wrap(err, "couldn't open an interactive dialog"))
	}

	return s.locOut(params, &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "apps.command.install.fillDialog",
			Other: "please continue by filling out the interactive form",
		},
	})
}
