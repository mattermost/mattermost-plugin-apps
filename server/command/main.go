// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type params struct {
	pluginContext *plugin.Context
	commandArgs   *model.CommandArgs
	current       []string
}

func (s *service) handleMain(in *params) (*model.CommandResponse, error) {
	subcommands := map[string]func(*params) (*model.CommandResponse, error){
		"info":                s.executeInfo,
		"install":             s.executeInstall,
		"debug-install-hello": s.executeDebugInstallHello,
		"debug-clean":         s.executeDebugClean,
		"debug-bindings":      s.executeDebugBindings,
		"debug-embedded":      s.executeDebugEmbedded,
		"debug-open-dialog":   s.executeDebugDialog,
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
		URL: s.apps.Configurator.GetConfig().PluginURL + "/hello/form/create_embedded",
		Context: &api.Context{
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

func (s *service) executeDebugDialog(params *params) (*model.CommandResponse, error) {
	if len(params.current) != 3 {
		return normalOut(params, nil, errors.New("not enough parameters"))
	}
	appID := params.current[0]
	url := params.current[1]
	dialogID := params.current[2]
	dialog, err := s.apps.Client.GetDebugDialog(api.AppID(appID), url, params.commandArgs.UserId, dialogID)
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
