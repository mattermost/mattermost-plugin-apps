// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func form(submit apps.Call) *apps.Form {
	return &apps.Form{
		Submit: &submit,
	}
}

func appIDForm(submit apps.Call) *apps.Form {
	f := form(submit)
	f.Fields = []apps.Field{
		{
			Name:                 fAppID,
			Type:                 apps.FieldTypeDynamicSelect,
			Description:          "select an App",
			Label:                fAppID,
			AutocompleteHint:     "App ID",
			AutocompletePosition: 1,
			IsRequired:           true,
		},
	}
	return f
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
				Label: string(app.Manifest.DisplayName),
			})
		}
	}
	return options, nil
}
