// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"github.com/mattermost/mattermost-plugin-cloudapps/server/utils/md"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

type params struct {
	pluginContext *plugin.Context
	commandArgs   *model.CommandArgs
	current       []string
}

func (s *service) handleMain(in *params) (*model.CommandResponse, error) {
	subcommands := map[string]func(*params) (*model.CommandResponse, error){
		"info":        s.executeInfo,
		"install":     s.executeInstall,
		"debug-clean": s.executeDebugClean,
	}

	return runSubcommand(subcommands, in)
}

func runSubcommand(
	subcommands map[string]func(*params) (*model.CommandResponse, error),
	params *params,
) (*model.CommandResponse, error) {
	if len(params.current) == 0 {
		return normalOut(params, md.MD("<><> TODO usage"),
			errors.New("expected a (sub-)command"))
	}
	if params.current[0] == "help" {
		return normalOut(params, md.MD("<><> TODO usage"), nil)
	}

	f := subcommands[params.current[0]]
	if f == nil {
		return normalOut(params, md.MD("<><> TODO usage"),
			errors.Errorf("unknown command: %s", params.current[0]))
	}

	p := *params
	p.current = params.current[1:]
	return f(&p)
}

func (s *service) executeDebugClean(params *params) (*model.CommandResponse, error) {
	return normalOut(params, md.MD("<><> TODO"), nil)
}
