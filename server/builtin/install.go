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

var installListedCall = apps.Call{
	Path: pInstallListed,
	Expand: &apps.Expand{
		AdminAccessToken: apps.ExpandAll, // ensure sysadmin
	},
}

func (a *builtinApp) installCommandBinding() apps.Binding {
	if a.conf.Get().MattermostCloudMode {
		return apps.Binding{
			Label:       "install",
			Location:    "install",
			Hint:        "[app ID]",
			Description: "Installs an App from the Marketplace",
			Call:        &installListedCall,
			Form:        appIDForm(installListedCall),
		}
	} else {
		return apps.Binding{
			Label:       "install",
			Location:    "install",
			Hint:        "[ listed | url ]",
			Description: "Installs an App, locally deployed or from a remote URL",
			Bindings: []apps.Binding{
				{
					Label:       "listed",
					Location:    "listed",
					Hint:        "[app ID]",
					Description: "Installs a listed App that has been locally deployed. (in the future, applicable Marketplace Apps will also be listed here).",
					Call:        &installListedCall,
					Form:        appIDForm(installListedCall),
				},
				{
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
				},
			},
		}
	}
}

func (a *builtinApp) installListed() handler {
	return handler{
		requireSysadmin: true,

		lookupf: func(creq apps.CallRequest) ([]apps.SelectOption, error) {
			res, err := a.lookupAppID(creq, func(app apps.ListedApp) bool {
				return !app.Installed
			})
			return res, err
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			appID := apps.AppID(creq.GetValue(fAppID, ""))
			m, err := a.proxy.GetManifest(appID)
			if err != nil {
				return apps.NewErrorCallResponse(err)
			}

			return formResponse(*a.newInstallConsentForm(*m, creq))
		},
	}
}

func (a *builtinApp) installURL() handler {
	return handler{
		requireSysadmin: true,

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
			_, err = a.proxy.StoreLocalManifest(*m)
			if err != nil {
				return apps.NewErrorCallResponse(err)
			}

			return formResponse(*a.newInstallConsentForm(*m, creq))
		},
	}
}
