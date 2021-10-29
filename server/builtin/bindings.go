package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func (a *builtinApp) bindings(creq apps.CallRequest) apps.CallResponse {
	loc := i18n.NewLocalizer(a.conf.I18N().Bundle, creq.Context.Locale)
	return apps.NewDataResponse(a.getBindings(creq, loc))
}

func (a *builtinApp) getBindings(creq apps.CallRequest, loc *i18n.Localizer) []apps.Binding {
	commands := []apps.Binding{
		a.infoCommandBinding(loc),
	}

	if utils.EnsureSysAdmin(a.conf.MattermostAPI(), creq.Context.ActingUserID) == nil {
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
					Label:    "apps",
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
