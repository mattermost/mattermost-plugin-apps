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
				Label:       a.conf.Local(loc, "command.debug.kv.edit.label"),
				Description: a.conf.Local(loc, "command.debug.kv.edit.description"),
				Hint:        a.conf.Local(loc, "command.debug.kv.edit.hint"),
				Call:        &debugKVEditCall,
				Form: a.appIDForm(debugKVEditCall, loc,
					apps.Field{
						Name:             fBase64Key,
						Type:             apps.FieldTypeText,
						Label:            a.conf.Local(loc, "field.kv.base64key.label"),
						Description:      a.conf.Local(loc, "field.kv.base64key.description"),
						AutocompleteHint: a.conf.Local(loc, "field.kv.base64key.hint"),
					},
					a.debugNamespaceField(loc),
					a.debugIDField(loc),
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
