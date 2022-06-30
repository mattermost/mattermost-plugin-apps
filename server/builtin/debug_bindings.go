// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
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

func (a *builtinApp) debugBindings(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	appID := apps.AppID(creq.GetValue(FieldAppID, ""))
	var bindings []apps.Binding
	out := ""
	var err error
	if appID == "" {
		bindings, err = a.proxy.GetBindings(r, creq.Context)
	} else {
		appRequest := r.WithDestination(appID)
		bindings, err = a.proxy.InvokeGetBindings(appRequest, creq.Context)
	}
	if err != nil {
		out += "### PROBLEMS:\n" + err.Error() + "\n\n"
	}
	out += utils.JSONBlock(bindings)
	return apps.NewTextResponse(out)
}
