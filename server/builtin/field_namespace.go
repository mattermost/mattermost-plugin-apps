// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
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
		DynamicSelectLookup: newUserCall(pLookupNamespace),
	}
}

func (a *builtinApp) lookupNamespace(creq apps.CallRequest) apps.CallResponse {
	if creq.SelectedField != fNamespace {
		return apps.NewErrorResponse(errors.Errorf("unknown field %q", creq.SelectedField))
	}
	appID := apps.AppID(creq.GetValue(fAppID, ""))
	if appID == "" {
		return apps.NewErrorResponse(errors.Errorf("please select --" + fAppID + " first"))
	}

	var options []apps.SelectOption
	_, namespaces, err := a.debugListKeys(appID)
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	for ns, c := range namespaces {
		options = append(options, apps.SelectOption{
			Value: ns,
			Label: fmt.Sprintf("%q (%v keys)", ns, c),
		})
	}
	return apps.NewLookupResponse(options)
}
