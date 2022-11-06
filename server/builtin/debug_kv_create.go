// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (a *builtinApp) debugKVCreateCommandBinding(loc *i18n.Localizer) apps.Binding {
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
		Form: &apps.Form{
			Submit: newUserCall(pDebugKVCreate),
			Fields: []apps.Field{
				a.appIDField(LookupInstalledApps, 1, true, loc),
				a.debugIDField(loc),
				a.namespaceField(0, false, loc),
			},
		},
	}
}

func (a *builtinApp) debugKVCreate(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	appID := apps.AppID(creq.GetValue(FieldAppID, ""))
	namespace := creq.GetValue(fNamespace, "")
	id := creq.GetValue(fID, "")

	appservicesRequest := r.WithSourceAppID(appID)
	data, err := a.appservices.KVGet(appservicesRequest, namespace, id)
	if err != nil && errors.Cause(err) != utils.ErrNotFound {
		return apps.NewErrorResponse(err)
	}
	if len(data) > 0 {
		return apps.NewErrorResponse(errors.New("key already exists, please use `/apps debug kv edit"))
	}

	_, err = a.appservices.KVSet(appservicesRequest, namespace, id, []byte("{}"))
	if err != nil {
		return apps.NewErrorResponse(err)
	}

	key := ""
	key, err = store.Hashkey(store.KVAppPrefix, appID, creq.Context.ActingUser.Id, namespace, id)
	if err != nil {
		return apps.NewErrorResponse(err)
	}

	creq.State = key
	return a.debugKVEditModalForm(r, creq)
}
