// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (a *builtinApp) appIDForm(call apps.Call, loc *i18n.Localizer) *apps.Form {
	return &apps.Form{
		Fields: []apps.Field{
			{
				Name: fAppID,
				Type: apps.FieldTypeDynamicSelect,
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "form.appIDForm.appID.description",
					Other: "select an App",
				}),
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "form.appIDForm.appID.label",
					Other: "app",
				}),
				AutocompleteHint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "form.appIDForm.appID.autocompleteHint",
					Other: "App ID",
				}),
				AutocompletePosition: 1,
				IsRequired:           true,
			},
		},
		Call: &call,
	}
}

func (a *builtinApp) lookupAppID(creq apps.CallRequest, includef func(apps.ListedApp) bool) ([]apps.SelectOption, error) {
	if creq.SelectedField != fAppID {
		return nil, errors.Errorf("unknown field %q", creq.SelectedField)
	}

	var options []apps.SelectOption
	marketplaceApps := a.proxy.GetListedApps(creq.Query, false)
	for _, app := range marketplaceApps {
		if includef == nil || includef(app) {
			options = append(options, apps.SelectOption{
				Value: string(app.Manifest.AppID),
				Label: app.Manifest.DisplayName,
			})
		}
	}
	return options, nil
}
