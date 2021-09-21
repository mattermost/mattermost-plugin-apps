// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (s *service) executeEnable(params *commandParams) (*model.CommandResponse, error) {
	if len(params.current) == 0 {
		return errorOut(params, errors.New("you need to specify the app id"))
	}

	appID := apps.AppID(params.current[0])

	cc := s.conf.Get().SetContextDefaultsForApp(appID, s.newCommandContext(params.commandArgs))

	out, err := s.proxy.EnableApp(s.newCommandIncoming(params.commandArgs), cc, appID)
	if err != nil {
		return errorOut(params, err)
	}

	return &model.CommandResponse{
		Text:         out,
		ResponseType: model.CommandResponseTypeEphemeral,
	}, nil
}

func (s *service) executeDisable(params *commandParams) (*model.CommandResponse, error) {
	if len(params.current) == 0 {
		return errorOut(params, errors.New("you need to specify the app id"))
	}

	appID := apps.AppID(params.current[0])

	cc := s.conf.Get().SetContextDefaultsForApp(appID, s.newCommandContext(params.commandArgs))

	out, err := s.proxy.DisableApp(s.newCommandIncoming(params.commandArgs), cc, appID)
	if err != nil {
		return errorOut(params, err)
	}

	return &model.CommandResponse{
		Text:         out,
		ResponseType: model.CommandResponseTypeEphemeral,
	}, nil
}
