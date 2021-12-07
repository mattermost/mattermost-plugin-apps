// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (a *builtinApp) debugBindingsCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Location: "bindings",
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.bindings.label",
			Other: "bindings",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.bindings.description",
			Other: "Display all bindings for the current context",
		}),
		Form: &apps.Form{
			Submit: newUserCall(pDebugBindings),
			Fields: []apps.Field{
				a.appIDField(LookupInstalledApps, 1, true, loc),
			},
		},
	}
}

func (a *builtinApp) debugBindings(creq apps.CallRequest) apps.CallResponse {
	appID := apps.AppID(creq.GetValue(fAppID, ""))
	var bindings []apps.Binding
	out := ""
	if appID == "" {
		var err error
		bindings, err = a.proxy.GetBindings(proxy.NewIncomingFromContext(creq.Context), creq.Context)
		if err != nil {
			return apps.NewErrorResponse(err)
		}
	} else {
		app, err := a.proxy.GetInstalledApp(appID)
		if err != nil {
			return apps.NewErrorResponse(err)
		}
		bindings, err = a.proxy.GetAppBindings(proxy.NewIncomingFromContext(creq.Context), creq.Context, *app)
		if err != nil {
			out += "\n\n### PROBLEMS:\n" + err.Error()
		}
		out += "\n\n"
	}

	out += utils.JSONBlock(bindings)
	return apps.NewTextResponse(out)
}
