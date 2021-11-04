// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"
)

var debugKVCreateCall = apps.Call{
	Path: pDebugKVCreate,
	Expand: &apps.Expand{
		ActingUser: apps.ExpandSummary,
	},
}

func (a *builtinApp) debugKVCreate() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func(loc *i18n.Localizer) apps.Binding {
			idF := a.debugIDField(loc)
			idF.IsRequired = true

			return apps.Binding{
				Location: "create",
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.kv.create.label",
					Other: "create",
				}),
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.kv.create.description",
					Other: "Create a new key-value for an App.",
				}),
				Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.kv.create.hint",
					Other: "[ AppID keyspec ]",
				}),
				Call: &debugKVCreateCall,
				Form: a.appIDForm(debugKVCreateCall, loc,
					a.debugNamespaceField(loc),
					idF,
				),
			}
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			appID := apps.AppID(creq.GetValue(fAppID, ""))
			namespace := creq.GetValue(fNamespace, "")
			id := creq.GetValue(fID, "")

			app, err := a.proxy.GetInstalledApp(appID)
			if err != nil {
				return apps.NewErrorResponse(err)
			}

			buf := []byte{}
			err = a.appservices.KVGet(app.BotUserID, namespace, id, &buf)
			if err != nil && errors.Cause(err) != utils.ErrNotFound {
				return apps.NewErrorResponse(err)
			}
			if len(buf) > 0 {
				return apps.NewErrorResponse(errors.New("Key already exists, please use `/apps debug kv edit"))
			}

			_, err = a.appservices.KVSet(app.BotUserID, namespace, id, []byte("{}"))
			if err != nil {
				return apps.NewErrorResponse(err)
			}

			key := ""
			key, err = store.Hashkey(config.KVAppPrefix, app.BotUserID, namespace, id)
			if err != nil {
				return apps.NewErrorResponse(err)
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
