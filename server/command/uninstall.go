// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (s *service) executeUninstall(params *commandParams) (*model.CommandResponse, error) {
	if len(params.current) == 0 {
		return errorOut(params, errors.New("you need to specify the app id"))
	}

	id := params.current[0]

	err := s.proxy.UninstallApp(params.commandArgs.Session.Id, params.commandArgs.UserId, apps.AppID(id))
	if err != nil {
		return errorOut(params, err)
	}

	return &model.CommandResponse{
		Text:         fmt.Sprintf("Uninstalled %s.", id),
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, nil
}
