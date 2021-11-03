// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"strconv"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var debugKVCleanCall = apps.Call{
	Path: pDebugKVClean,
	Expand: &apps.Expand{
		ActingUser: apps.ExpandSummary,
	},
}

func (a *builtinApp) debugKVClean() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func(loc *i18n.Localizer) apps.Binding {
			return apps.Binding{
				Location: "clean",
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.kv.clean.label",
					Other: "clean",
				}),
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.kv.clean.description",
					Other: "Delete KV keys for an app, in a specific namespace.",
				}),
				Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.kv.clean.hint",
					Other: "[ App ID ]",
				}),
				Call: &debugKVInfoCall,
				Form: a.appIDForm(debugKVListCall, loc, a.debugNamespaceField(loc)),
			}
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			appID := apps.AppID(creq.GetValue(fAppID, ""))
			namespace := creq.GetValue(fNamespace, "")
			app, err := a.proxy.GetInstalledApp(appID)
			if err != nil {
				return apps.NewErrorResponse(err)
			}

			n := 0
			err = a.appservices.KVList(app.BotUserID, namespace, func(key string) error {
				n++
				return a.conf.MattermostAPI().KV.Delete(key)
			})
			if err != nil {
				return apps.NewErrorResponse(err)
			}

			loc := a.newLocalizer(creq)
			return apps.NewTextResponse(a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "command.debug.kv.clean.submit",
					Other: "Deleted {{.Count}} keys for `{{.AppID}}`, namespace `{{.Namespace}}`.",
				},
				TemplateData: map[string]string{
					"Count":     strconv.Itoa(n),
					"AppID":     string(appID),
					"Namespace": namespace,
				},
			}))
		},

		lookupf: func(creq apps.CallRequest) ([]apps.SelectOption, error) {
			return a.lookupAppID(creq, func(app apps.ListedApp) bool {
				return app.Installed
			})
		},
	}
}
