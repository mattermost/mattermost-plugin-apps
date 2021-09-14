// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var installURLCall = apps.Call{
	Path: pInstallURL,
	Expand: &apps.Expand{
		AdminAccessToken: apps.ExpandAll, // ensure sysadmin
	},
}

func (a *builtinApp) installURL() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func() apps.Binding {
			return apps.Binding{
				Label:       "url",
				Location:    "url",
				Hint:        "[manifest.json URL]",
				Description: "Installs an App from an HTTP URL",
				Call:        &installURLCall,
				Form: &apps.Form{
					Fields: []apps.Field{
						{
							Name:                 fURL,
							Type:                 apps.FieldTypeText,
							Description:          "enter the URL for the app's manifest.json",
							Label:                fURL,
							AutocompleteHint:     "URL",
							AutocompletePosition: 1,
							IsRequired:           true,
						},
					},
					Call: &installURLCall,
				},
			}
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
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
		},
	}
}
