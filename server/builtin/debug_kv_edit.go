// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"encoding/base64"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/nicksnyder/go-i18n/v2/i18n"
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

		commandBinding: func(loc *i18n.Localizer) apps.Binding {
			return apps.Binding{
				Location:    "edit",
				Label:       "edit",                                     // <>/<> Localize
				Description: "View or edit specific KV keys of an app.", // <>/<> Localize
				Call:        &debugKVEditCall,
				Form: a.appIDForm(debugKVEditCall, loc,
					apps.Field{
						Name:             fBase64Key,
						Label:            "base64-key", // <>/<> Localize
						Type:             apps.FieldTypeText,
						Description:      "base64-encoded key, from `debug kv list`. No other flags needed.", // <>/<> Localize
						AutocompleteHint: "[ key ]",                                                          // <>/<> Localize
					},
					apps.Field{
						Name:        fNamespace,
						Label:       fNamespace, // <>/<> Localize
						Type:        apps.FieldTypeDynamicSelect,
						Description: "App-specific namespace (up to 2 letters). Requires `--app` and `--id`.", // <>/<> Localize
					},
					apps.Field{
						Name:             fID,
						Label:            fID, // <>/<> Localize
						Type:             apps.FieldTypeText,
						Description:      "App-specific ID, any length.", // <>/<> Localize
						AutocompleteHint: "[ id ]",                       // <>/<> Localize
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

		lookupf: a.debugAppNamespaceLookup,
	}
}
