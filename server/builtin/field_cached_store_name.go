// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (a *builtinApp) cachedStoreNameField(loc *i18n.Localizer) apps.Field {
	return apps.Field{
		Name: FieldAppID,
		Type: apps.FieldTypeDynamicSelect,
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.cached_store_name.description",
			Other: "Select a cached store",
		}),
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.cached_store_name.label",
			Other: "cached-store",
		}),
		SelectDynamicLookup: newUserCall(pLookupCachedStoreName),
	}
}

func (a *builtinApp) lookupCachedStoreName(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	var options []apps.SelectOption
	info, _ := a.appservices.KVDebugInfo(r)
	for name := range info.CachedStoreCountByName {
		options = append(options, apps.SelectOption{
			Value: name,
			Label: name,
		})
	}
	return apps.NewLookupResponse(options)
}
