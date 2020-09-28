// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/http/dialog"
)

func (s *service) executeInstall(params *params) (*model.CommandResponse, error) {
	manifestURL := ""
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&manifestURL, "url", "", "manifest URL")

	err := fs.Parse(params.current)
	if err != nil {
		return normalOut(params, nil, err)
	}

	manifest, err := s.apps.Client.GetManifest(manifestURL)
	if err != nil {
		return normalOut(params, nil, err)
	}

	conf := s.apps.Configurator.GetConfig()

	// Finish the installation when the Dialog is submitted, see
	// <plugin>/http/dialog/install.go
	err = s.apps.Mattermost.Frontend.OpenInteractiveDialog(
		dialog.NewInstallAppDialog(manifest, conf.PluginURL, params.commandArgs))
	if err != nil {
		return normalOut(params, nil, errors.Wrap(err, "couldn't open an interactive dialog"))
	}

	team, err := s.apps.Mattermost.Team.Get(params.commandArgs.TeamId)
	if err != nil {
		return normalOut(params, nil, err)
	}

	return &model.CommandResponse{
		GotoLocation: params.commandArgs.SiteURL + "/" + team.Name + "/messages/@" + constants.BotUserName,
		Text:         fmt.Sprintf("redirected to the DM with @%s to continue installing **%s**", constants.BotUserName, manifest.DisplayName),
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, nil
}
