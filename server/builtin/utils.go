// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func appIDForm(call apps.Call, extraFields ...apps.Field) *apps.Form {
	form := &apps.Form{
		Fields: []apps.Field{
			{
				Name:                 fAppID,
				Label:                fAppID,
				Type:                 apps.FieldTypeDynamicSelect,
				Description:          "Select an App.",
				AutocompleteHint:     "App ID",
				AutocompletePosition: 1,
				IsRequired:           true,
			},
		},
		Call: &call,
	}
	form.Fields = append(form.Fields, extraFields...)
	return form
}

func (a *builtinApp) lookupAppID(creq apps.CallRequest, includef func(apps.ListedApp) bool) ([]apps.SelectOption, error) {
	if creq.SelectedField != fAppID {
		return nil, errors.Errorf("unknown field %q", creq.SelectedField)
	}

	var options []apps.SelectOption
	marketplaceApps := a.proxy.GetListedApps(creq.Query, true)
	for _, app := range marketplaceApps {
		if includef == nil || includef(app) {
			options = append(options, apps.SelectOption{
				Value: string(app.Manifest.AppID),
				Label: string(app.Manifest.DisplayName),
			})
		}
	}
	return options, nil
}
