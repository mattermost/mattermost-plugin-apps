// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

func (s *service) executeUninstall(params *commandParams) (*model.CommandResponse, error) {
	client, err := s.newMMClient(params.commandArgs)
	if err != nil {
		return errorOut(params, err)
	}

	cmd := strings.Fields(params.commandArgs.Command)
	if len(cmd) != 3 {
		return errorOut(params, errors.New("incorrect number of uninstall parameters"))
	}

	appID := apps.AppID(cmd[len(cmd)-1])

	cc := s.conf.GetConfig().SetContextDefaultsForApp(appID, s.newCommandContext(params.commandArgs))

	out, err := s.proxy.UninstallApp(client, params.commandArgs.Session.Id, cc, appID)
	if err != nil {
		return errorOut(params, err)
	}

	return &model.CommandResponse{
		Text:         string(out),
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, nil
}
