// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func blankForm(submit *apps.Call) *apps.Form {
	return &apps.Form{
		Submit: submit,
	}
}

func appIDForm(submit, lookup *apps.Call) *apps.Form {
	f := blankForm(submit)
	f.Fields = []apps.Field{
		{
			Name:                 fAppID,
			Type:                 apps.FieldTypeDynamicSelect,
			Description:          "select an App",
			Label:                fAppID,
			AutocompleteHint:     "App ID",
			AutocompletePosition: 1,
			IsRequired:           true,
			SelectLookup:         lookup,
		},
	}
	return f
}

func (a *builtinApp) lookupAppID(creq apps.CallRequest, includef func(apps.ListedApp) bool) apps.CallResponse {
	if creq.SelectedField != fAppID {
		return apps.NewErrorResponse(errors.Errorf("unknown field %q", creq.SelectedField))
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
	return apps.NewLookupResponse(options)
}

func newAdminCall(path string) *apps.Call {
	call := apps.NewCall(path)
	call.Expand = &apps.Expand{
		AdminAccessToken: apps.ExpandAll,
	}
	return call
}
