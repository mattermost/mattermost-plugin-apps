// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var installURLCall = apps.Call{
	Path: pInstallURL,
}

var installURLForm = apps.Form{
	Fields: []apps.Field{
		{
			Name:                 fURL,
			Type:                 apps.FieldTypeText,
			Description:          "enter the URL for the app's manifest.json",
			Label:                fURL,
			AutocompleteHint:     "URL",
			AutocompletePosition: 1,
			IsRequired: true,
		},
	},
	Call: &installURLCall,
}

func (a *builtinApp) installURLSubmit(creq apps.CallRequest) apps.CallResponse {
	manifestURL := creq.GetValue(fURL, "")
	conf := a.conf.Get()
	data, err := a.httpOut.GetFromURL(manifestURL, conf.DeveloperMode, apps.MaxManifestSize)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	m, err := apps.DecodeCompatibleManifest(data)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	return a.installCommandSubmit(*m, creq)
}
