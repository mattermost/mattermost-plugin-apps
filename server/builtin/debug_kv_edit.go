// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"encoding/base64"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

var debugKVEditCall = apps.Call{
	Path: pDebugKVEdit,
	Expand: &apps.Expand{
		ActingUser: apps.ExpandSummary,
	},
}

func (a *builtinApp) debugKVEdit() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func() apps.Binding {
			return apps.Binding{
				Label:       "edit",
				Location:    "edit",
				Hint:        "",
				Description: "View or edit specific KV keys of an app.",
				Call:        &debugKVEditCall,
				Form: appIDForm(debugKVEditCall,
					apps.Field{
						Name:             fBase64Key,
						Label:            "base64-key",
						Type:             apps.FieldTypeText,
						Description:      "base64-encoded key, from `debug kv list`. No other flags needed.",
						AutocompleteHint: "[ key ]",
					},
					apps.Field{
						Name:        fNamespace,
						Label:       fNamespace,
						Type:        apps.FieldTypeText,
						Description: "App-specific namespace (up to 2 letters). Requires `--app` and `--id`.",
					},
					apps.Field{
						Name:             fID,
						Label:            fID,
						Type:             apps.FieldTypeText,
						Description:      "App-specific ID, any length.",
						AutocompleteHint: "[ id ]",
					},
				),
			}
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			appID := apps.AppID(creq.GetValue(fAppID, ""))
			base64Key := creq.GetValue(fBase64Key, "")
			namespace := creq.GetValue(fNamespace, "")
			id := creq.GetValue(fID, "")

			key := ""
			if base64Key != "" {
				decoded, err := base64.URLEncoding.DecodeString(base64Key)
				if err != nil {
					return apps.NewErrorResponse(err)
				}
				key = string(decoded)
			} else {
				app, err := a.proxy.GetInstalledApp(appID)
				if err != nil {
					return apps.NewErrorResponse(err)
				}
				key, err = store.Hashkey(config.KVAppPrefix, app.BotUserID, namespace, id)
				if err != nil {
					return apps.NewErrorResponse(err)
				}
			}

			creq.State = key
			form, err := a.debugKVEditModal().formf(creq)
			if err != nil {
				return apps.NewErrorResponse(err)
			}
			return apps.NewFormResponse(*form)
		},

		lookupf: func(creq apps.CallRequest) ([]apps.SelectOption, error) {
			return a.lookupAppID(creq, func(app apps.ListedApp) bool {
				return app.Installed
			})
		},
	}
}
