package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var noParameters = apps.Form{
	// TODO translate this?
	Title: "command with no parameters",
}

func (a *builtinApp) bindings(creq apps.CallRequest) apps.CallResponse {
	loc := a.newLocalizer(creq)
	return apps.NewDataResponse(a.getBindings(creq, loc))
}

func (a *builtinApp) getBindings(creq apps.CallRequest, loc *i18n.Localizer) []apps.Binding {
	commands := []apps.Binding{
		a.info().commandBinding(loc),
	}

	if creq.Context.ActingUser != nil && creq.Context.ActingUser.IsSystemAdmin() {
		commands = append(commands,
			a.debugCommandBinding(loc),
			a.disable().commandBinding(loc),
			a.enable().commandBinding(loc),
			a.installCommandBinding(loc),
			a.list().commandBinding(loc),
			a.uninstall().commandBinding(loc),
		)
	}

	return []apps.Binding{
		{
			Location: apps.LocationCommand,
			Bindings: []apps.Binding{
				{
					Label:    "apps", //  "/apps" in all locales
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
