package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

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

	if utils.EnsureSysAdmin(a.mm, creq.Context.ActingUserID) == nil {
		commands = append(commands,
			a.installCommandBinding(),
			a.uninstallCommandBinding(),
			a.enableCommandBinding(),
			a.disableCommandBinding(),
			a.listCommandBinding(),
			a.debugCommandBinding(),
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

func commandBinding(label, path, hint, descr string) apps.Binding {
	return apps.Binding{
		Label:       label,
		Location:    apps.Location(label),
		Hint:        hint,
		Description: descr,
		Call: &apps.Call{
			Path: path,
		},
	}
}
