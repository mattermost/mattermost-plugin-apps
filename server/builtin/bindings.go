package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (a *builtinApp) bindings(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	loc := a.newLocalizer(creq)
	return apps.NewDataResponse(a.getBindings(r, creq, loc))
}

func (a *builtinApp) getBindings(r *incoming.Request, creq apps.CallRequest, loc *i18n.Localizer) []apps.Binding {
	commands := []apps.Binding{
		a.infoCommandBinding(loc),
	}

	if creq.Context.ActingUser != nil && creq.Context.ActingUser.IsSystemAdmin() {
		if r.Config.Get().DeveloperMode {
			commands = append(commands, a.debugCommandBinding(loc))
		}
		commands = append(commands,
			a.disableCommandBinding(loc),
			a.enableCommandBinding(loc),
			a.installCommandBinding(r, loc),
			a.listCommandBinding(loc),
			a.uninstallCommandBinding(loc),
			a.settingsCommandBinding(loc),
		)
	}

	return []apps.Binding{
		{
			Location: apps.LocationCommand,
			Bindings: []apps.Binding{
				{
					Label:    "apps", // "/apps" in all locales
					Location: "apps",
					Description: r.API.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "command.base.description",
						Other: "Mattermost Apps",
					}),
					Bindings: commands,
				},
			},
		},
	}
}
