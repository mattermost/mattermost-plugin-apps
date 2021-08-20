package command

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/server/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Service interface {
	config.Configurable
	ExecuteCommand(pluginContext *plugin.Context, commandArgs *model.CommandArgs) (*model.CommandResponse, error)
}

type service struct {
	conf    config.Service
	proxy   proxy.Service
	httpOut httpout.Service
}

var _ Service = (*service)(nil)

type commandParams struct {
	pluginContext *plugin.Context
	commandArgs   *model.CommandArgs
	current       []string
}

type commandHandler struct {
	f            func(*commandParams) (*model.CommandResponse, error)
	subCommands  map[string]commandHandler
	devOnly      bool
	autoComplete *model.AutocompleteData
}

func (s *service) allSubCommands(conf config.Config) map[string]commandHandler {
	uninstallAC := model.NewAutocompleteData("uninstall", "", "Uninstall an app")
	uninstallAC.AddTextArgument("ID of the app to uninstall", "appID", "")
	uninstallAC.RoleID = model.SYSTEM_ADMIN_ROLE_ID

	enableAC := model.NewAutocompleteData("enable", "", "Enable an app")
	enableAC.AddTextArgument("ID of the app to enable", "appID", "")
	enableAC.RoleID = model.SYSTEM_ADMIN_ROLE_ID

	disenableAC := model.NewAutocompleteData("disable", "", "Disable an app")
	disenableAC.AddTextArgument("ID of the app to disable", "appID", "")
	disenableAC.RoleID = model.SYSTEM_ADMIN_ROLE_ID

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
		"enable": {
			f:            s.checkSystemAdmin(s.executeEnable),
			autoComplete: enableAC,
		},
		"disable": {
			f:            s.checkSystemAdmin(s.executeDisable),
			autoComplete: disenableAC,
		},
	}

	if conf.DeveloperMode {
		debugAddManifestAC := model.NewAutocompleteData("debug-add-manifest", "", "Add a manifest to the local list of known apps")
		debugAddManifestAC.AddNamedTextArgument("url", "URL of the manifest to add", "URL", "", true)
		debugAddManifestAC.RoleID = model.SYSTEM_ADMIN_ROLE_ID

		debugCleanAC := model.NewAutocompleteData("debug-clean", "", "Delete all KV data")
		debugCleanAC.RoleID = model.SYSTEM_ADMIN_ROLE_ID

		all["debug-bindings"] = commandHandler{
			f:            s.executeDebugBindings,
			devOnly:      true,
			autoComplete: model.NewAutocompleteData("debug-bindings", "", "List bindings"),
		}
		all["debug-clean"] = commandHandler{
			f:            s.checkSystemAdmin(s.executeDebugClean),
			devOnly:      true,
			autoComplete: debugCleanAC,
		}
		// TODO ticket: change to watch-manifest
		all["debug-add-manifest"] = commandHandler{
			f:            s.checkSystemAdmin(s.executeDebugAddManifest),
			devOnly:      true,
			autoComplete: debugAddManifestAC,
		}
	}

	all["install"] = s.installCommand(conf)

	return all
}

func (s *service) installCommand(conf config.Config) commandHandler {
	h := commandHandler{
		autoComplete: &model.AutocompleteData{
			Trigger:  "install",
			HelpText: "Install an App.",
			RoleID:   model.SYSTEM_ADMIN_ROLE_ID,
		},
		subCommands: map[string]commandHandler{},
	}

	if conf.MattermostCloudMode {
		// install only by ID (from the marketplace) in cloud mode
		installMarketplaceAC := model.NewAutocompleteData("marketplace", "", "Install an App from the Mattermost Marketplace")
		installMarketplaceAC.Arguments = append(installMarketplaceAC.Arguments, &model.AutocompleteArg{
			HelpText: "ID of the app to install",
			Type:     model.AutocompleteArgTypeText,
			Data: &model.AutocompleteTextArg{
				Hint: "App ID",
			},
			Required: true,
		})

		h.subCommands[installMarketplaceAC.Trigger] = commandHandler{
			f:            s.checkSystemAdmin(s.executeInstallMarketplace),
			autoComplete: installMarketplaceAC,
		}
	} else {
		installHTTPAC := model.NewAutocompleteData("http", "", "Install an App running as a HTTP server")
		// install from URL in the on-prem mode
		installHTTPAC.Arguments = append(installHTTPAC.Arguments, &model.AutocompleteArg{
			Name:     "",
			HelpText: "URL of the App's manifest",
			Type:     model.AutocompleteArgTypeText,
			Data: &model.AutocompleteTextArg{
				Hint: "URL",
			},
			Required: true,
		})

		installHTTPAC.Arguments = append(installHTTPAC.Arguments, &model.AutocompleteArg{
			Name:     "app-secret",
			HelpText: "(HTTP) App's JWT secret used to authenticate incoming messages from Mattermost.",
			Type:     model.AutocompleteArgTypeText,
			Data: &model.AutocompleteTextArg{
				Hint: "Secret string",
			},
			Required: false,
		})
		h.subCommands[installHTTPAC.Trigger] = commandHandler{
			f:            s.checkSystemAdmin(s.executeInstallHTTP),
			autoComplete: installHTTPAC,
		}

		installAWSAC := model.NewAutocompleteData("aws", "", "Install an App running as an AWS lambda function")
		installAWSAC.Arguments = append(installAWSAC.Arguments, &model.AutocompleteArg{
			HelpText: "ID of the app to install",
			Type:     model.AutocompleteArgTypeText,
			Data: &model.AutocompleteTextArg{
				Hint: "App ID",
			},
			Required: true,
		})
		installAWSAC.Arguments = append(installAWSAC.Arguments, &model.AutocompleteArg{
			HelpText: "Version of the app to install",
			Type:     model.AutocompleteArgTypeText,
			Data: &model.AutocompleteTextArg{
				Hint: "version",
			},
			Required: true,
		})

		h.subCommands[installAWSAC.Trigger] = commandHandler{
			f:            s.checkSystemAdmin(s.executeInstallAWS),
			autoComplete: installAWSAC,
		}
	}

	return h
}

func MakeService(configService config.Service, proxy proxy.Service, httpOut httpout.Service) (Service, error) {
	s := &service{
		conf:    configService,
		proxy:   proxy,
		httpOut: httpOut,
	}
	conf := s.conf.Get()

	err := s.registerCommand(conf)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *service) Configure(conf config.Config) {
	err := s.registerCommand(conf)
	if err != nil {
		s.conf.Logger().WithError(err).Warnf("Failed to re-register command")
	}
}

func (s *service) registerCommand(conf config.Config) error {
	subCommands := s.allSubCommands(conf)
	var subs []string
	for t := range subCommands {
		subs = append(subs, t)
	}
	sort.Strings(subs)
	helpText := "Available commands: " + strings.Join(subs, ", ")

	// Add autocomplete for the subcommands alphabetically
	ac := model.NewAutocompleteData(config.CommandTrigger, "[command]", helpText)

	AddACForSubCommands(subCommands, ac)

	err := s.conf.MattermostAPI().SlashCommand.Register(&model.Command{
		Trigger:          config.CommandTrigger,
		AutoComplete:     true,
		AutoCompleteDesc: "Manage Apps",
		AutoCompleteHint: fmt.Sprintf("Usage: `/%s info`.", config.CommandTrigger),
		AutocompleteData: ac,
	})

	return err
}

func AddACForSubCommands(subCommands map[string]commandHandler, rootAC *model.AutocompleteData) {
	var subs []string
	for t := range subCommands {
		subs = append(subs, t)
	}
	sort.Strings(subs)

	for _, t := range subs {
		if len(subCommands[t].subCommands) > 0 {
			AddACForSubCommands(subCommands[t].subCommands, subCommands[t].autoComplete)
		}
		rootAC.AddCommand(subCommands[t].autoComplete)
	}
}

// Handle should be called by the plugin when a command invocation is received from the Mattermost server.
func (s *service) ExecuteCommand(pluginContext *plugin.Context, commandArgs *model.CommandArgs) (*model.CommandResponse, error) {
	params := &commandParams{
		pluginContext: pluginContext,
		commandArgs:   commandArgs,
	}
	if pluginContext == nil || commandArgs == nil {
		return s.errorOut(params, utils.NewLocError(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.error.invalidArguments",
				Other: "invalid arguments to command.Handler. Please contact your system administrator",
			},
		}), errors.New("invalid arguments to command.Handler. Please contact your system administrator"))
	}

	conf := s.conf.MattermostConfig().Config()
	enableOAuthServiceProvider := conf.ServiceSettings.EnableOAuthServiceProvider
	if enableOAuthServiceProvider == nil || !*enableOAuthServiceProvider {
		url := commandArgs.SiteURL + "/admin_console/integrations/integration_management"
		return s.errorOut(params, utils.NewLocError(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.enableOAuth",
				Other: "the system setting `Enable OAuth 2.0 Service Provider` needs to be enabled in order for the Apps plugin to work. Please go to {{.URL}} and enable it.",
			},
			TemplateData: map[string]string{
				"URL": url,
			},
		}), fmt.Errorf("the system setting `Enable OAuth 2.0 Service Provider` needs to be enabled in order for the Apps plugin to work. Please go to %s and enable it.", url))
	}

	enableBotAccounts := conf.ServiceSettings.EnableBotAccountCreation
	if enableBotAccounts == nil || !*enableBotAccounts {
		url := commandArgs.SiteURL + "/admin_console/integrations/bot_accounts"
		return s.errorOut(params, utils.NewLocError(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.enableBot",
				Other: "the system setting `Enable Bot Account Creation` needs to be enabled in order for the Apps plugin to work. Please go to {{.URL}} and enable it.",
			},
			TemplateData: map[string]string{
				"URL": url,
			},
		}), fmt.Errorf("the system setting `Enable Bot Account Creation` needs to be enabled in order for the Apps plugin to work. Please go to %s and enable it.", url))
	}

	split := strings.Fields(commandArgs.Command)
	if len(split) < 2 {
		return s.errorOut(params, utils.NewLocError(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.error.noSubCommand",
				Other: "no subcommand specified, nothing to do",
			},
		}), errors.New("no subcommand specified, nothing to do"))
	}

	command := split[0]
	if command != "/"+config.CommandTrigger {
		return s.errorOut(params, utils.NewLocError(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.notSupported",
				Other: "{{.Command}} is not a supported command and should not have been invoked. Please contact your system administrator",
			},
			TemplateData: map[string]string{
				"Command": command,
			},
		}), fmt.Errorf("%s is not a supported command and should not have been invoked. Please contact your system administrator", command))
	}

	params.current = split[1:]

	return s.handleMain(params)
}

func (s *service) handleMain(in *commandParams) (*model.CommandResponse, error) {
	conf := s.conf.Get()
	return s.runSubcommand(s.allSubCommands(conf), in)
}

func (s *service) runSubcommand(subcommands map[string]commandHandler, params *commandParams) (*model.CommandResponse, error) {
	if len(params.current) == 0 {
		return s.errorOut(params, utils.NewLocError(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.error.needSubCommand",
				Other: "expected a (sub-)command",
			},
		}), errors.New("expected a (sub-)command"))
	}
	if params.current[0] == "help" {
		return out(params, "TODO usage")
	}

	c, ok := subcommands[params.current[0]]
	if !ok {
		command := params.current[0]
		return s.errorOut(params, utils.NewLocError(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.unknown",
				Other: "unknown command: {{.Command}}",
			},
			TemplateData: map[string]string{
				"Command": command,
			},
		}), fmt.Errorf("unknown command: %s", command))
	}

	conf := s.conf.Get()
	if c.devOnly && !conf.DeveloperMode {
		command := params.current[0]
		return s.errorOut(params, utils.NewLocError(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.developersOnly",
				Other: "{{.Command}} is only available in developers mode. You need to enable `Developer Mode` and `Testing Commands` in the System Console.",
			},
			TemplateData: map[string]string{
				"Command": command,
			},
		}), fmt.Errorf("%s is only available in developers mode. You need to enable `Developer Mode` and `Testing Commands` in the System Console.", command))
	}

	p := *params
	p.current = params.current[1:]

	if len(c.subCommands) > 0 {
		return s.runSubcommand(c.subCommands, &p)
	}

	return c.f(&p)
}

func (s *service) checkSystemAdmin(handler func(*commandParams) (*model.CommandResponse, error)) func(*commandParams) (*model.CommandResponse, error) {
	return func(p *commandParams) (*model.CommandResponse, error) {
		if !s.conf.MattermostAPI().User.HasPermissionTo(p.commandArgs.UserId, model.PERMISSION_MANAGE_SYSTEM) {
			return s.errorOut(p, utils.NewLocError(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "apps.command.error.mustBeAdmin",
					Other: "you need to be a system admin to run this command",
				},
			}), errors.New("you need to be a system admin to run this command"))
		}

		return handler(p)
	}
}

func (s *service) newCommandContext(commandArgs *model.CommandArgs) *apps.Context {
	return s.conf.Get().SetContextDefaults(&apps.Context{
		UserAgentContext: apps.UserAgentContext{
			TeamID:    commandArgs.TeamId,
			ChannelID: commandArgs.ChannelId,
		},
		ActingUserID: commandArgs.UserId,
		UserID:       commandArgs.UserId,
	})
}

func (s *service) newMMClient(commandArgs *model.CommandArgs) (mmclient.Client, utils.LocError, error) {
	return mmclient.NewHTTPClient(s.conf, commandArgs.Session.Id, commandArgs.UserId)
}

func out(params *commandParams, out string) (*model.CommandResponse, error) {
	txt := utils.CodeBlock(params.commandArgs.Command+"\n") + out

	return &model.CommandResponse{
		Text:         txt,
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, nil
}

func (s *service) errorOut(params *commandParams, locError utils.LocError, err error) (*model.CommandResponse, error) {
	bundle := s.conf.I18N()
	loc := bundle.GetUserLocalizer(params.commandArgs.UserId)
	out := utils.CodeBlock(params.commandArgs.Command)
	out += bundle.LocalizeWithConfig(loc, &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "apps.command.error.header",
			Other: "Command failed.",
		},
	})
	if locError != nil {
		locErrorString := locError.Error(bundle, loc)
		locErrorString = bundle.LocalizeWithConfig(loc, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.error.localized",
				Other: "Error: **{{.LocalizedError}}**",
			},
			TemplateData: map[string]string{
				"LocalizedError": locErrorString,
			},
		})
		out += "\n" + locErrorString
	}
	if err != nil {
		out += "\n" + bundle.LocalizeWithConfig(loc, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "apps.command.error.details",
				Other: "Error details:",
			},
		})
		out += "\n" + utils.CodeBlock(err.Error())
	}

	return &model.CommandResponse{
		Text:         out,
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, err
}

func (s *service) locOut(params *commandParams, locText *i18n.LocalizeConfig) (*model.CommandResponse, error) {
	loc := s.conf.I18N().GetUserLocalizer(params.commandArgs.UserId)
	locString := s.conf.I18N().LocalizeWithConfig(loc, locText)
	txt := utils.CodeBlock(params.commandArgs.Command+"\n") + locString

	return &model.CommandResponse{
		Text:         txt,
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, nil
}
