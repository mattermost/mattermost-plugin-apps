// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (a *builtinApp) namespaceField(pos int, isRequired bool, loc *i18n.Localizer) apps.Field {
	return apps.Field{
		Name: fNamespace,
		Type: apps.FieldTypeDynamicSelect,
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.namespace.label",
			Other: "namespace",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.namespace.description",
			Other: "App-specific namespace (up to 2 letters). See `debug kv info` for the list of app's namespaces.",
		}),
		IsRequired:           isRequired,
		AutocompletePosition: pos,
		AutocompleteHint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.namespace.hint",
			Other: "namespace (up to 2 letters)",
		}),
		SelectDynamicLookup: newUserCall(pLookupNamespace),
	}
}

func (a *builtinApp) lookupNamespace(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	appID := apps.AppID(creq.GetValue(FieldAppID, ""))
	if appID == "" {
		return apps.NewErrorResponse(errors.Errorf("please select --" + FieldAppID + " first"))
	}

	var options []apps.SelectOption
	appInfo, err := a.appservices.KVDebugAppInfo(r, appID)
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	for ns, c := range appInfo.AppKVCountByNamespace {
		options = append(options, apps.SelectOption{
			Value: ns,
			Label: fmt.Sprintf("%q (%v keys)", ns, c),
		})
	}
	return apps.NewLookupResponse(options)
}
