package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

var noParameters = apps.Form{
	Title: "command with no parameters",
}

func (a *builtinApp) getBindings(creq apps.CallRequest) apps.CallResponse {
	return apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Data: a.bindings(creq),
	}
}

func (a *builtinApp) bindings(creq apps.CallRequest) []apps.Binding {
	commands := []apps.Binding{
		a.infoCommandBinding(),
	}

	if utils.EnsureSysAdmin(a.conf.MattermostAPI(), creq.Context.ActingUserID) == nil {
		commands = append(commands,
			a.debugCommandBinding(),
			a.disableCommandBinding(),
			a.enableCommandBinding(),
			a.installCommandBinding(),
			a.listCommandBinding(),
			a.uninstallCommandBinding(),
		)
	}

	return []apps.Binding{
		{
			Location: apps.LocationCommand,
			Bindings: []apps.Binding{
				{
					Label:       "apps",
					Location:    "apps",
					Description: "Mattermost Apps",
					Bindings:    commands,
				},
			},
		},
	}
}
