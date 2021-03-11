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

type commandHandler struct {
	f     func(*params) (*model.CommandResponse, error)
	debug bool
}

func (s *service) handleMain(in *params, developerMode bool) (*model.CommandResponse, error) {
	subcommands := map[string]commandHandler{
		"debug-bindings":     {s.executeDebugBindings, true},
		"debug-clean":        {s.executeDebugClean, true},
		"debug-install-http": {s.executeDebugInstallHTTPHello, true},
		"debug-install-aws":  {s.executeDebugInstallAWSHello, true},
		"info":               {s.executeInfo, false},
		"list":               {s.executeList, false},
		"install":            {s.executeInstall, false},
		"uninstall":          {s.checkSystemAdmin(s.executeUninstall), false},
	}

	return runSubcommand(subcommands, in, developerMode)
}

func runSubcommand(subcommands map[string]commandHandler, params *params, developerMode bool) (*model.CommandResponse, error) {
	if len(params.current) == 0 {
		return errorOut(params, errors.New("expected a (sub-)command"))
	}
	if params.current[0] == "help" {
		return out(params, md.MD("TODO usage"))
	}

	c, ok := subcommands[params.current[0]]
	if !ok {
		return errorOut(params, errors.Errorf("unknown command: %s", params.current[0]))
	}

	if c.debug && !developerMode {
		return errorOut(params, errors.Errorf("%s is only available in developers mode. You need to enable `Developer Mode` and `Testing Commands` in the System Console.", params.current[0]))
	}

	p := *params
	p.current = params.current[1:]
	return c.f(&p)
}

func (s *service) checkSystemAdmin(handler func(*params) (*model.CommandResponse, error)) func(*params) (*model.CommandResponse, error) {
	return func(p *params) (*model.CommandResponse, error) {
		if !s.api.Mattermost.User.HasPermissionTo(p.commandArgs.UserId, model.PERMISSION_MANAGE_SYSTEM) {
			return errorOut(p, errors.New("you need to be a system admin to run this command"))
		}

		return handler(p)
	}
}
