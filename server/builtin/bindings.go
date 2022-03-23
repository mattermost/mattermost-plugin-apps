package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (a *builtinApp) bindings(_ *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	loc := a.newLocalizer(creq)
	return apps.NewDataResponse(a.getBindings(creq, loc))
}

func (a *builtinApp) getBindings(creq apps.CallRequest, loc *i18n.Localizer) []apps.Binding {
	commands := []apps.Binding{
		a.infoCommandBinding(loc),
	}

	if creq.Context.ActingUser != nil && creq.Context.ActingUser.IsSystemAdmin() {
		commands = append(commands,
			a.debugCommandBinding(loc),
			a.disableCommandBinding(loc),
			a.enableCommandBinding(loc),
			a.installCommandBinding(loc),
			a.listCommandBinding(loc),
			a.uninstallCommandBinding(loc),
		)
	}

	return []apps.Binding{
		{
			Location: apps.LocationCommand,
			Bindings: []apps.Binding{
				{
					Label:    "apps", // "/apps" in all locales
					Location: "apps",
					Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "command.base.description",
						Other: "Mattermost Apps",
					}),
					Bindings: commands,
				},
			},
		},
	}
}
