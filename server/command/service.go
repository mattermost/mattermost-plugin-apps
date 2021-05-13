package command

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type Service interface {
	ExecuteCommand(pluginContext *plugin.Context, commandArgs *model.CommandArgs) (*model.CommandResponse, error)
}

type service struct {
	mm      *pluginapi.Client
	conf    config.Service
	proxy   proxy.Service
	httpOut httpout.Service
}

var _ Service = (*service)(nil)

type params struct {
	pluginContext *plugin.Context
	commandArgs   *model.CommandArgs
	current       []string
}

type commandHandler struct {
	f            func(*params) (*model.CommandResponse, error)
	devOnly      bool
	autoComplete *model.AutocompleteData
}

func (s *service) allCommands() map[string]commandHandler {
	uninstallAC := model.NewAutocompleteData("uninstall", "", "Uninstall an app")
	uninstallAC.AddTextArgument("ID of the app to uninstall", "appID", "")

	all := map[string]commandHandler{
		"info": {
			f:            s.executeInfo,
			autoComplete: model.NewAutocompleteData("info", "", "Display debugging information"),
		},
		"list": {
			f:            s.executeList,
			autoComplete: model.NewAutocompleteData("list", "", "List installed and registered apps"),
		},
		"uninstall": {
			f:            s.checkSystemAdmin(s.executeUninstall),
			autoComplete: uninstallAC,
		},
	}

	conf := s.conf.GetConfig()

	if conf.DeveloperMode {
		debugAddManifestAC := model.NewAutocompleteData("debug-add-manifest", "", "Add a manifest to the local list of known apps")
		debugAddManifestAC.AddNamedTextArgument("url", "URL of the manifest to add", "URL", "", true)

		all["debug-bindings"] = commandHandler{
			f:            s.executeDebugBindings,
			devOnly:      true,
			autoComplete: model.NewAutocompleteData("debug-bindings", "", "List bindings"),
		}
		all["debug-clean"] = commandHandler{
			f:            s.executeDebugClean,
			devOnly:      true,
			autoComplete: model.NewAutocompleteData("debug-clean", "", "Delete all KV data"),
		}
		// TODO ticket: change to watch-manifest
		all["debug-add-manifest"] = commandHandler{
			f:            s.executeDebugAddManifest,
			devOnly:      true,
			autoComplete: debugAddManifestAC,
		}
	}

	all["install"] = s.installCommand(conf)

	return all
}

func (s *service) installCommand(conf config.Config) commandHandler {
	h := commandHandler{
		f: s.checkSystemAdmin(s.executeInstall),
		autoComplete: &model.AutocompleteData{
			Trigger:  "install",
			HelpText: "Install an App.",
			RoleID:   model.SYSTEM_ADMIN_ROLE_ID,
		},
	}

	if conf.MattermostCloudMode {
		// install only by ID (from the marketplace) in cloud mode
		h.autoComplete.Arguments = append(h.autoComplete.Arguments, &model.AutocompleteArg{
			HelpText: "ID of the app to install",
			Type:     model.AutocompleteArgTypeText,
			Data: &model.AutocompleteTextArg{
				Hint: "secret string",
			},
			Required: true,
		})
	} else {
		// install from URL in the on-prem mode
		h.autoComplete.Arguments = append(h.autoComplete.Arguments, &model.AutocompleteArg{
			Name:     "url",
			HelpText: "URL of the App's manifest",
			Type:     model.AutocompleteArgTypeText,
			Data: &model.AutocompleteTextArg{
				Hint: "URL",
			},
			Required: true,
		})
	}

	h.autoComplete.Arguments = append(h.autoComplete.Arguments, &model.AutocompleteArg{
		Name:     "app-secret",
		HelpText: "(HTTP) App's JWT secret used to authenticate incoming messages from Mattermost.",
		Type:     model.AutocompleteArgTypeText,
		Data: &model.AutocompleteTextArg{
			Hint: "secret string",
		},
	})

	return h
}

func MakeService(mm *pluginapi.Client, configService config.Service, proxy proxy.Service, httpOut httpout.Service) (Service, error) {
	s := &service{
		mm:      mm,
		conf:    configService,
		proxy:   proxy,
		httpOut: httpOut,
	}
	subCommands := s.allCommands()
	var subs []string
	for t := range subCommands {
		subs = append(subs, t)
	}
	sort.Strings(subs)
	helpText := "Available commands: " + strings.Join(subs, ", ")

	// Add autocomplete for the subcommands alphabetically
	ac := model.NewAutocompleteData(config.CommandTrigger, "[command]", helpText)
	for _, t := range subs {
		ac.AddCommand(subCommands[t].autoComplete)
	}

	err := mm.SlashCommand.Register(&model.Command{
		Trigger:          config.CommandTrigger,
		AutoComplete:     true,
		AutoCompleteDesc: "Manage Apps",
		AutoCompleteHint: fmt.Sprintf("Usage: `/%s info`.", config.CommandTrigger),
		AutocompleteData: ac,
	})
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Handle should be called by the plugin when a command invocation is received from the Mattermost server.
func (s *service) ExecuteCommand(pluginContext *plugin.Context, commandArgs *model.CommandArgs) (*model.CommandResponse, error) {
	params := &params{
		pluginContext: pluginContext,
		commandArgs:   commandArgs,
	}
	if pluginContext == nil || commandArgs == nil {
		return errorOut(params, errors.New("invalid arguments to command.Handler. Please contact your system administrator"))
	}

	conf := s.conf.GetMattermostConfig().Config()
	enableOAuthServiceProvider := conf.ServiceSettings.EnableOAuthServiceProvider
	if enableOAuthServiceProvider == nil || !*enableOAuthServiceProvider {
		return errorOut(params, errors.Errorf("the system setting `Enable OAuth 2.0 Service Provider` needs to be enabled in order for the Apps plugin to work. Please go to %s/admin_console/integrations/integration_management and enable it.", commandArgs.SiteURL))
	}

	enableBotAccounts := conf.ServiceSettings.EnableBotAccountCreation
	if enableBotAccounts == nil || !*enableBotAccounts {
		return errorOut(params, errors.Errorf("the system setting `Enable Bot Account Creation` needs to be enabled in order for the Apps plugin to work. Please go to %s/admin_console/integrations/bot_accounts and enable it.", commandArgs.SiteURL))
	}

	split := strings.Fields(commandArgs.Command)
	if len(split) < 2 {
		return errorOut(params, errors.New("no subcommand specified, nothing to do"))
	}

	command := split[0]
	if command != "/"+config.CommandTrigger {
		return errorOut(params, errors.Errorf("%q is not a supported command and should not have been invoked. Please contact your system administrator", command))
	}

	params.current = split[1:]

	return s.handleMain(params)
}

func (s *service) handleMain(in *params) (*model.CommandResponse, error) {
	return s.runSubcommand(s.allCommands(), in)
}

func (s *service) runSubcommand(subcommands map[string]commandHandler, params *params) (*model.CommandResponse, error) {
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

	conf := s.conf.GetConfig()
	if c.devOnly && !conf.DeveloperMode {
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
func out(params *params, out md.Markdowner) (*model.CommandResponse, error) {
	txt := md.CodeBlock(params.commandArgs.Command+"\n") + out.Markdown()
	return &model.CommandResponse{
		Text:         string(txt),
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, nil
}

func errorOut(params *params, err error) (*model.CommandResponse, error) {
	txt := md.CodeBlock(params.commandArgs.Command+"\n") +
		md.Markdownf("Command failed. Error: **%s**\n", err.Error())
	return &model.CommandResponse{
		Text:         string(txt),
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, err
}
