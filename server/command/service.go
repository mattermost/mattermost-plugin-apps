package command

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-cloudapps/server/apps"
	"github.com/mattermost/mattermost-plugin-cloudapps/server/constants"
	"github.com/mattermost/mattermost-plugin-cloudapps/server/utils/md"
)

type Service interface {
	ExecuteCommand(pluginContext *plugin.Context, commandArgs *model.CommandArgs) (*model.CommandResponse, error)
}

type service struct {
	apps *apps.Service
}

var _ Service = (*service)(nil)

func MakeService(apps *apps.Service) (Service, error) {
	conf := apps.Config.GetConfig()

	s := &service{
		apps: apps,
	}
	err := apps.Mattermost.SlashCommand.Register(&model.Command{
		Trigger:          constants.CommandTrigger,
		DisplayName:      conf.BuildConfig.Manifest.Name,
		Description:      conf.BuildConfig.Manifest.Description,
		AutoComplete:     true,
		AutoCompleteDesc: "Manage Cloud Apps",
		AutoCompleteHint: fmt.Sprintf("Usage: `/%s info`.", constants.CommandTrigger),
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
		return normalOut(params, nil,
			errors.New("invalid arguments to command.Handler. Please contact your system administrator"))
	}
	split := strings.Fields(commandArgs.Command)
	if len(split) < 2 {
		return normalOut(params, nil,
			errors.New("no subcommand specified, nothing to do"))
	}
	command := split[0]
	if command != "/"+constants.CommandTrigger {
		return normalOut(params, nil,
			errors.Errorf("%q is not a supported command and should not have been invoked. Please contact your system administrator", command))
	}
	params.current = split[1:]

	return s.handleMain(params)
}

func normalOut(params *params, out md.Markdowner, err error) (*model.CommandResponse, error) {
	message := md.CodeBlock(params.commandArgs.Command + "\n")
	if err != nil {
		message += md.Markdownf("Command failed. Error: **%s**\n", err.Error())
	} else {
		message += out.Markdown()
	}

	return &model.CommandResponse{
		Text:         message.String(),
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}, err
}
