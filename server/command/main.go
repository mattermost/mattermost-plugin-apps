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
	f            func(*params) (*model.CommandResponse, error)
	debug        bool
	autoComplete *model.AutocompleteData
}

func (s *service) getSubCommands() map[string]commandHandler {
	debugAddManifestAC := model.NewAutocompleteData("debug-add-manifest", "", "Add a manifest to the local list of known apps")
	debugAddManifestAC.AddNamedTextArgument("url", "URL of the manifest to add", "URL", "", true)

	installAC := model.NewAutocompleteData("install", "", "Install a registered app")
	installAC.AddNamedTextArgument("app-id", "ID of the app to install", "appID", "", true)
	installAC.AddNamedTextArgument("app-secret", "Secret used to secure connection to App", "App Secret", "", false)

	uninstallAC := model.NewAutocompleteData("uninstall", "", "Uninstall an app")
	uninstallAC.AddTextArgument("ID of the app to uninstall", "appID", "")

	return map[string]commandHandler{
		"debug-bindings":     {s.executeDebugBindings, true, model.NewAutocompleteData("debug-bindings", "", "List bindings")},
		"debug-clean":        {s.executeDebugClean, true, model.NewAutocompleteData("debug-clean", "", "Delete all KV data")},
		"debug-add-manifest": {s.executeDebugAddManifest, true, debugAddManifestAC},
		"info":               {s.executeInfo, false, model.NewAutocompleteData("info", "", "Display debugging information")},
		"list":               {s.executeList, false, model.NewAutocompleteData("list", "", "List installed and registered apps")},
		"install":            {s.executeInstall, false, installAC},
		"uninstall":          {s.checkSystemAdmin(s.executeUninstall), false, uninstallAC},
	}
}

func (s *service) handleMain(in *params, developerMode bool) (*model.CommandResponse, error) {
	return runSubcommand(s.getSubCommands(), in, developerMode)
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
		if !s.mm.User.HasPermissionTo(p.commandArgs.UserId, model.PERMISSION_MANAGE_SYSTEM) {
			return errorOut(p, errors.New("you need to be a system admin to run this command"))
		}

		return handler(p)
	}
}
