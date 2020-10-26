// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type params struct {
	pluginContext *plugin.Context
	commandArgs   *model.CommandArgs
	current       []string
}

func (s *service) handleMain(in *params) (*model.CommandResponse, error) {
	subcommands := map[string]func(*params) (*model.CommandResponse, error){
		"info":    s.executeInfo,
		"install": s.executeInstall,
		// For Debug
		"debug-clean":     s.executeDebugClean,
		"debug-locations": s.executeDebugLocations,
		"debug-embedded":  s.executeDebugEmbedded,
		// For internal use only
		"openDialog": s.openDialog,
	}

	return runSubcommand(subcommands, in)
}

func runSubcommand(
	subcommands map[string]func(*params) (*model.CommandResponse, error),
	params *params,
) (*model.CommandResponse, error) {
	if len(params.current) == 0 {
		return normalOut(params, md.MD("TODO usage"),
			errors.New("expected a (sub-)command"))
	}
	if params.current[0] == "help" {
		return normalOut(params, md.MD("TODO usage"), nil)
	}

	f := subcommands[params.current[0]]
	if f == nil {
		return normalOut(params, md.MD("TODO usage"),
			errors.Errorf("unknown command: %s", params.current[0]))
	}

	p := *params
	p.current = params.current[1:]
	return f(&p)
}

func (s *service) executeDebugClean(params *params) (*model.CommandResponse, error) {
	return normalOut(params, md.MD("TODO"), nil)
}

func (s *service) executeDebugLocations(params *params) (*model.CommandResponse, error) {
	locations, err := s.apps.API.GetLocations(params.commandArgs.UserId, params.commandArgs.ChannelId)
	if err != nil {
		return normalOut(params, md.MD("error"), err)
	}
	return normalOut(params, md.JSONBlock(locations), nil)
}

func (s *service) executeDebugEmbedded(params *params) (*model.CommandResponse, error) {
	_, err := s.apps.Client.PostCall(&apps.Call{
		FormURL: s.apps.Configurator.GetConfig().PluginURL + "/hello/wish/create_embedded",
		Context: &apps.Context{
			AppID:        "hello",
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

func (s *service) openDialog(params *params) (*model.CommandResponse, error) {
	if len(params.current) != 3 {
		return normalOut(params, nil, errors.New("not enough parameters"))
	}
	appID := params.current[0]
	url := params.current[1]
	dialogID := params.current[2]
	dialog, err := s.apps.Client.GetDialog(store.AppID(appID), url, params.commandArgs.UserId, dialogID)
	if err != nil {
		return normalOut(params, nil, err)
	}

	dialog.TriggerId = params.commandArgs.TriggerId

	err = s.apps.Mattermost.Frontend.OpenInteractiveDialog(*dialog)
	if err != nil {
		return normalOut(params, nil, err)
	}

	return normalOut(params, md.MD(""), nil)
}
