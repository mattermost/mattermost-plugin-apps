// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

var debugKVInfoCall = apps.Call{
	Path: pDebugKVInfo,
}

func (a *builtinApp) debugKVInfo() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func() apps.Binding {
			return apps.Binding{
				Label:       "info",
				Location:    "info",
				Hint:        "[ AppID ]",
				Description: "Display KV store statistics for an app.",
				Call:        &debugKVInfoCall,
				Form:        appIDForm(debugKVInfoCall),
			}
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			appID := apps.AppID(creq.GetValue(fAppID, ""))
			app, err := a.proxy.GetInstalledApp(appID)
			if err != nil {
				return apps.NewErrorResponse(err)
			}

			n := 0
			namespaces := map[string]int{}
			err = a.appservices.KVList(app.BotUserID, "", func(key string) error {
				_, _, ns, _, e := store.ParseHashkey(key)
				if e != nil {
					return e
				}
				namespaces[ns] = namespaces[ns] + 1
				n++
				return nil
			})
			if err != nil {
				return apps.NewErrorResponse(err)
			}

			message := fmt.Sprintf("%v total keys for `%s`.\n", n, appID)

			if len(namespaces) > 0 {
				message += "\nNamespaces:\n"
			}
			for ns, c := range namespaces {
				if ns == "" {
					ns = "(none)"
				}
				message += fmt.Sprintf("  - `%s`: %v\n", ns, c)
			}
			return apps.NewTextResponse(message)
		},

		lookupf: func(creq apps.CallRequest) ([]apps.SelectOption, error) {
			return a.lookupAppID(creq, nil)
		},
	}
}
