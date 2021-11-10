// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (a *builtinApp) appIDForm(submitCall *apps.Call, lookupCall *apps.Call, loc *i18n.Localizer, extraFields ...apps.Field) *apps.Form {
	form := &apps.Form{
		Fields: []apps.Field{
			{
				Name: fAppID,
				Type: apps.FieldTypeDynamicSelect,
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "field.appID.description",
					Other: "Select an App or enter the App ID",
				}),
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "field.appID.label",
					Other: "app",
				}),
				AutocompletePosition: 1,
				IsRequired:           true,
				SelectLookup:         lookupCall,
			},
		},
		Submit: submitCall,
	}
	form.Fields = append(form.Fields, extraFields...)
	return form
}

func (a *builtinApp) lookupAppID(creq apps.CallRequest, includef func(apps.ListedApp) bool) apps.CallResponse {
	if creq.SelectedField != fAppID {
		return apps.NewErrorResponse(errors.Errorf("unknown field %q", creq.SelectedField))
	}

	var options []apps.SelectOption
	marketplaceApps := a.proxy.GetListedApps(creq.Query, true)
	for _, app := range marketplaceApps {
		if includef == nil || includef(app) {
			options = append(options, apps.SelectOption{
				Value: string(app.Manifest.AppID),
				Label: app.Manifest.DisplayName,
			})
		}
	}
	return apps.NewLookupResponse(options)
}

func newAdminCall(path string) *apps.Call {
	call := apps.NewCall(path)
	call.Expand = &apps.Expand{
		ActingUser:            apps.ExpandSummary,
		ActingUserAccessToken: apps.ExpandAll,
	}
	return call
}
