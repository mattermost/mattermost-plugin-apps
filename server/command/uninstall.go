// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (s *service) executeUninstall(params *commandParams) (*model.CommandResponse, error) {
	if len(params.current) == 0 {
		return s.errorOut(params, utils.NewLocError(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.uninstall.error.needAppId",
				Other: "you need to specify the app id",
			},
		}), errors.New("you need to specify the app id"))
	}

	client, locErr, err := s.newMMClient(params.commandArgs)
	if err != nil {
		return s.errorOut(params, locErr, err)
	}

	appID := apps.AppID(params.current[0])

	cc := s.conf.Get().SetContextDefaultsForApp(appID, s.newCommandContext(params.commandArgs))

	out, err := s.proxy.UninstallApp(client, params.commandArgs.Session.Id, cc, appID)
	if err != nil {
		return s.errorOut(params, nil, err)
	}

	return &model.CommandResponse{
		Text:         out,
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, nil
}
