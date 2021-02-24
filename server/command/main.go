// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type params struct {
	pluginContext *plugin.Context
	commandArgs   *model.CommandArgs
	current       []string
}

func (s *service) handleMain(in *params) (*model.CommandResponse, error) {
	subcommands := map[string]func(*params) (*model.CommandResponse, error){
		"debug-bindings":        s.executeDebugBindings,
		"debug-clean":           s.executeDebugClean,
		"debug-install-builtin": s.executeDebugInstallBuiltinHello,
		"debug-install-http":    s.executeDebugInstallHTTPHello,
		"debug-install-aws":     s.executeDebugInstallAWSHello,
		"info":                  s.executeInfo,
		"list":                  s.executeList,
		"install":               s.executeInstall,
		"uninstall":             s.checkSystemAdmin(s.executeUninstall),
	}

	return runSubcommand(subcommands, in)
}

func runSubcommand(
	subcommands map[string]func(*params) (*model.CommandResponse, error),
	params *params,
) (*model.CommandResponse, error) {
	if len(params.current) == 0 {
		return errorOut(params, errors.New("expected a (sub-)command"))
	}
	if params.current[0] == "help" {
		return out(params, md.MD("TODO usage"))
	}

	f := subcommands[params.current[0]]
	if f == nil {
		return errorOut(params, errors.Errorf("unknown command: %s", params.current[0]))
	}

	p := *params
	p.current = params.current[1:]
	return f(&p)
}

func (s *service) checkSystemAdmin(handler func(*params) (*model.CommandResponse, error)) func(*params) (*model.CommandResponse, error) {
	return func(p *params) (*model.CommandResponse, error) {
		if !s.api.Mattermost.User.HasPermissionTo(p.commandArgs.UserId, model.PERMISSION_MANAGE_SYSTEM) {
			return errorOut(p, errors.New("you need to be a system admin to run this command"))
		}

		return handler(p)
	}
}
