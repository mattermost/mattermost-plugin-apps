package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/nicksnyder/go-i18n/v2/i18n"
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

	if utils.EnsureSysAdmin(a.conf.MattermostAPI(), creq.Context.ActingUserID) == nil {
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
					Label:       "apps", //  "/apps" in all locales
					Location:    "apps",
					Description: a.conf.Local(loc, "command.base.description"),
					Bindings:    commands,
				},
			},
		},
	}
}
