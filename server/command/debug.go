// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/http/helloapp"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (s *service) executeDebugClean(params *params) (*model.CommandResponse, error) {
	_ = s.apps.Mattermost.KV.DeleteAll()
	return normalOut(params, md.MD("Deleted all KV records"), nil)
}

func (s *service) executeDebugBindings(params *params) (*model.CommandResponse, error) {
	bindings, err := s.apps.API.GetBindings(&api.Context{
		ActingUserID: params.commandArgs.UserId,
		UserID:       params.commandArgs.UserId,
		TeamID:       params.commandArgs.TeamId,
		ChannelID:    params.commandArgs.ChannelId,
	})
	if err != nil {
		return normalOut(params, md.MD("error"), err)
	}
	return normalOut(params, md.JSONBlock(bindings), nil)
}

func (s *service) executeDebugEmbedded(params *params) (*model.CommandResponse, error) {
	_, err := s.apps.API.Call(&api.Call{
		URL: s.apps.Configurator.GetConfig().PluginURL + constants.HelloAppPath + helloapp.PathSendSurvey,
		Context: &api.Context{
			AppID:        helloapp.AppID,
			ActingUserID: params.commandArgs.UserId,
			ChannelID:    params.commandArgs.ChannelId,
			TeamID:       params.commandArgs.TeamId,
			UserID:       params.commandArgs.UserId,
		},
	})

	if err != nil {
		return normalOut(params, nil, err)
	}

	return normalOut(params, md.MD("The app will send you the form"), nil)
}

func (s *service) executeDebugInstallHello(params *params) (*model.CommandResponse, error) {
	params.current = []string{
		"--app-secret", helloapp.AppSecret,
		"--url", s.apps.Configurator.GetConfig().PluginURL + constants.HelloAppPath + helloapp.PathManifest,
		"--force",
	}
	return s.executeInstall(params)
}
