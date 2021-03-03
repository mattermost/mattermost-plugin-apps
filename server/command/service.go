package command

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type Service interface {
	ExecuteCommand(pluginContext *plugin.Context, commandArgs *model.CommandArgs) (*model.CommandResponse, error)
}

type service struct {
	mm    *pluginapi.Client
	conf  api.Configurator
	proxy api.Proxy
	admin api.Admin
}

var _ Service = (*service)(nil)

func MakeService(mm *pluginapi.Client, configService api.Configurator, proxy api.Proxy, admin api.Admin) (Service, error) {
	s := &service{
		mm:    mm,
		conf:  configService,
		proxy: proxy,
		admin: admin,
	}

	conf := configService.GetConfig()
	err := mm.SlashCommand.Register(&model.Command{
		Trigger:          api.CommandTrigger,
		DisplayName:      conf.BuildConfig.Manifest.Name,
		Description:      conf.BuildConfig.Manifest.Description,
		AutoComplete:     true,
		AutoCompleteDesc: "Manage Cloud Apps",
		AutoCompleteHint: fmt.Sprintf("Usage: `/%s info`.", api.CommandTrigger),
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

	enableOAuthServiceProvider := s.mm.Configuration.GetConfig().ServiceSettings.EnableOAuthServiceProvider
	if enableOAuthServiceProvider == nil || !*enableOAuthServiceProvider {
		return errorOut(params, errors.Errorf("the system setting `Enable OAuth 2.0 Service Provider` need to be enabled in order for the Apps plugin to work. Please go to %s/admin_console/integrations/integration_management and enable it.", commandArgs.SiteURL))
	}

	split := strings.Fields(commandArgs.Command)
	if len(split) < 2 {
		return errorOut(params, errors.New("no subcommand specified, nothing to do"))
	}

	command := split[0]
	if command != "/"+api.CommandTrigger {
		return errorOut(params, errors.Errorf("%q is not a supported command and should not have been invoked. Please contact your system administrator", command))
	}

	params.current = split[1:]

	return s.handleMain(params)
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
