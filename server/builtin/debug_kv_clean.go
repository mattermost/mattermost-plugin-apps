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
				Location:    "clean",
				Label:       a.conf.Local(loc, "command.debug.kv.clean.label"),
				Description: a.conf.Local(loc, "command.debug.kv.clean.description"),
				Hint:        a.conf.Local(loc, "command.debug.kv.clean.hint"),
				Call:        &debugKVInfoCall,
				Form:        a.appIDForm(debugKVListCall, loc, a.debugNamespaceField(loc)),
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

			return apps.NewTextResponse(a.conf.LocalWithTemplate(a.newLocalizer(creq),
				"command.debug.kv.clean.submit",
				map[string]string{
					"Count":     strconv.Itoa(n),
					"AppID":     string(appID),
					"Namespace": namespace,
				}))
		},

		lookupf: func(creq apps.CallRequest) ([]apps.SelectOption, error) {
			return a.lookupAppID(creq, func(app apps.ListedApp) bool {
				return app.Installed
			})
		},
	}
}
