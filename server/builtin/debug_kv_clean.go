// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var debugKVCleanCall = apps.Call{
	Path: pDebugKVClean,
}

func (a *builtinApp) debugKVClean() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func() apps.Binding {
			return apps.Binding{
				Label:       "clean",
				Location:    "clean",
				Hint:        "[ AppID ]",
				Description: "Deletes KV keys for an app, in a specific namespace.",
				Call:        &debugKVInfoCall,
				Form:        appIDForm(debugKVListCall, namespaceField),
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

			return apps.NewTextResponse(fmt.Sprintf("Deleted %v  keys for `%s`, namespace `%s`.\n", n, appID, namespace))
		},

		lookupf: func(creq apps.CallRequest) ([]apps.SelectOption, error) {
			return a.lookupAppID(creq, func(app apps.ListedApp) bool {
				return app.Installed
			})
		},
	}
}
