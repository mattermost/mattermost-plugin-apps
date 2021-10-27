// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"encoding/base64"
	"fmt"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var debugKVListCall = apps.Call{
	Path: pDebugKVList,
	Expand: &apps.Expand{
		ActingUser: apps.ExpandSummary,
	},
}

func (a *builtinApp) debugKVList() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func() apps.Binding {
			return apps.Binding{
				Label:       "list",
				Location:    "list",
				Hint:        "[ AppID ]",
				Description: "Display the list of KV keys for an app, in a specific namespace.",
				Call:        &debugKVInfoCall,
				Form:        appIDForm(debugKVListCall, namespaceField, base64Field),
			}
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			appID := apps.AppID(creq.GetValue(fAppID, ""))
			namespace := creq.GetValue(fNamespace, "")
			encode := creq.BoolValue(fBase64)
			app, err := a.proxy.GetInstalledApp(appID)
			if err != nil {
				return apps.NewErrorResponse(err)
			}

			keys := []string{}
			err = a.appservices.KVList(app.BotUserID, namespace, func(key string) error {
				keys = append(keys, key)
				return nil
			})
			if err != nil {
				return apps.NewErrorResponse(err)
			}

			message := fmt.Sprintf("%v total keys for `%s`, namespace `%s`\n", len(keys), appID, namespace)
			if encode {
				message += "**NOTE**: keys are base64-encoded for pasting into " +
					"`/apps debug kv edit` command. Use `/apps debug kv list --base64 false` " +
					"to output raw values.\n"
				for _, key := range keys {
					message += fmt.Sprintf("- `%s`\n", base64.URLEncoding.EncodeToString([]byte(key)))
				}
			} else {
				message += "```\n"
				for _, key := range keys {
					message += fmt.Sprintln(key)
				}
				message += "```\n"
			}
			return apps.NewTextResponse(message)
		},

		lookupf: func(creq apps.CallRequest) ([]apps.SelectOption, error) {
			return a.lookupAppID(creq, func(app apps.ListedApp) bool {
				return app.Installed
			})
		},
	}
}
