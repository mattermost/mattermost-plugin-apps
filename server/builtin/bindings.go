package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var noParameters = apps.Form{
	Title: "command with no parameters",
}

func (a *builtinApp) bindings(creq apps.CallRequest) apps.CallResponse {
	return apps.NewDataResponse(a.getBindings(creq))
}

func (a *builtinApp) getBindings(creq apps.CallRequest) []apps.Binding {
	commands := []apps.Binding{
		a.info().commandBinding(),
	}

	if creq.Context.ActingUser != nil && creq.Context.ActingUser.IsSystemAdmin() {
		commands = append(commands,
			a.debugCommandBinding(),
			a.disable().commandBinding(),
			a.enable().commandBinding(),
			a.installCommandBinding(),
			a.list().commandBinding(),
			a.uninstall().commandBinding(),
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
