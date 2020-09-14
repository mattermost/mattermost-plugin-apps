package command

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-cloudapps/server/configurator"
	"github.com/mattermost/mattermost-plugin-cloudapps/server/constants"
	"github.com/mattermost/mattermost-plugin-cloudapps/server/utils/md"
)

type Mattermost interface {
	RegisterCommand(*model.Command) error
}

type Command interface {
	Init(*configurator.BuildConfig) error
	Handle(pluginContext *plugin.Context, commandArgs *model.CommandArgs) (out md.MD, err error)
}

type command struct {
	Mattermost
	configurator.Configurator
}

var _ Command = (*command)(nil)

func NewCommand(configurator configurator.Configurator, mm Mattermost) Command {
	return &command{
		Mattermost:   mm,
		Configurator: configurator,
	}
}

// Register should be called by the plugin to register all necessary commands
func (c *command) Init(buildConfig *configurator.BuildConfig) error {
	return c.Mattermost.RegisterCommand(&model.Command{
		Trigger:          constants.CommandTrigger,
		DisplayName:      buildConfig.Manifest.Name,
		Description:      buildConfig.Manifest.Description,
		AutoComplete:     true,
		AutoCompleteDesc: "Manage Cloud Apps",
		AutoCompleteHint: fmt.Sprintf("Usage: `/%s info`.", constants.CommandTrigger),
	})
}

// Handle should be called by the plugin when a command invocation is received from the Mattermost server.
func (c *command) Handle(pluginContext *plugin.Context, commandArgs *model.CommandArgs) (out md.MD, err error) {
	defer func() {
		prefix := md.CodeBlock(commandArgs.Command) + "\n"
		if err != nil {
			prefix += md.Markdownf("Command failed. Error: **%s**\n", err.Error())
		}
		out = prefix + out
	}()

	if pluginContext == nil || commandArgs == nil {
		return "", errors.New("invalid arguments to command.Handler. Please contact your system administrator")
	}
	split := strings.Fields(commandArgs.Command)
	if len(split) < 2 {
		return "", errors.New("no subcommand specified, nothing to do")
	}
	command := split[0]
	if command != "/"+constants.CommandTrigger {
		return "", errors.Errorf("%q is not a supported command and should not have been invoked. Please contact your system administrator", command)
	}
	parameters := split[1:]

	return c.handleMain(parameters)
}

func runSubcommand(
	subcommands map[string]func([]string) (md.MD, error),
	parameters []string,
) (md.MD, error) {
	if len(parameters) == 0 {
		return "<><> TODO usage", errors.New("expected a (sub-)command")
	}
	if parameters[0] == "help" {
		return "<><> TODO usage", nil
	}

	f := subcommands[parameters[0]]
	if f == nil {
		return "<><> TODO usage", errors.Errorf("unknown command: %s", parameters[0])
	}

	return f(parameters[1:])
}

// func normalOut(out md.Markdowner, err error) (md.MD, error) {
// 	if err != nil {
// 		return "", err
// 	}
// 	return out.Markdown(), nil
// }
