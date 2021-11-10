// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"encoding/base64"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

func (a *builtinApp) debugKVEditCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Location: "edit",
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.kv.edit.label",
			Other: "edit",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.kv.edit.description",
			Other: "View or edit specific KV keys of an app.",
		}),
		Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.kv.edit.hint",
			Other: "[ AppID keyspec ]",
		}),
		Form: a.appIDForm(newAdminCall(pDebugKVEdit).WithLocale(), newAdminCall(pDebugKVEditLookup), loc,
			a.debugBase64KeyField(loc),
			a.debugNamespaceField(loc),
			a.debugIDField(loc),
		),
	}
}

func (a *builtinApp) debugKVEdit(creq apps.CallRequest) apps.CallResponse {
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
	return a.debugKVEditModalForm(creq)
}

func (a *builtinApp) debugKVEditLookup(creq apps.CallRequest) apps.CallResponse {
	return a.debugAppNamespaceLookup(creq)
}
