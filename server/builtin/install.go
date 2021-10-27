// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
)

var installHTTPCall = apps.Call{
	Path: pInstallHTTP,
	Expand: &apps.Expand{
		ActingUser: apps.ExpandSummary,
	},
}

var installListedCall = apps.Call{
	Path: pInstallListed,
	Expand: &apps.Expand{
		ActingUser: apps.ExpandSummary,
	},
}

func (a *builtinApp) installCommandBinding() apps.Binding {
	if a.conf.Get().MattermostCloudMode {
		return apps.Binding{
			Label:       "install",
			Location:    "install",
			Hint:        "[app ID]",
			Description: "Install an App from the Marketplace.",
			Call:        &installListedCall,
			Form:        appIDForm(installListedCall),
		}
	} else {
		return apps.Binding{
			Label:       "install",
			Location:    "install",
			Hint:        "[ listed | url ]",
			Description: "Install an App, locally deployed or from a remote URL.",
			Bindings: []apps.Binding{
				{
					Label:       "listed",
					Location:    "listed",
					Hint:        "[app ID]",
					Description: "Install a listed App that has been locally deployed. (in the future, applicable Marketplace Apps will also be listed here).",
					Call:        &installListedCall,
					Form:        appIDForm(installListedCall),
				},
				{
					Label:       "http",
					Location:    "http",
					Hint:        "[URL to manifest.json]",
					Description: "Install an HTTP App from a URL.",
					Call:        &installHTTPCall,
					Form: &apps.Form{
						Fields: []apps.Field{
							{
								Name:                 fURL,
								Label:                fURL,
								Type:                 apps.FieldTypeText,
								Description:          "enter the HTTP URL for the app's manifest.json.",
								AutocompleteHint:     "URL",
								AutocompletePosition: 1,
								IsRequired:           true,
							},
						},
						Call: &installHTTPCall,
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
			res, err := a.lookupAppID(creq, nil)
			return res, err
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			appID := apps.AppID(creq.GetValue(fAppID, ""))
			m, err := a.proxy.GetManifest(appID)
			if err != nil {
				return apps.NewErrorResponse(err)
			}

			return apps.NewFormResponse(*a.newInstallConsentForm(*m, creq, ""))
		},
	}
}

func (a *builtinApp) installHTTP() handler {
	return handler{
		requireSysadmin: true,

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			manifestURL := creq.GetValue(fURL, "")
			conf := a.conf.Get()
			data, err := a.httpOut.GetFromURL(manifestURL, conf.DeveloperMode, apps.MaxManifestSize)
			if err != nil {
				return apps.NewErrorResponse(err)
			}
			m, err := apps.DecodeCompatibleManifest(data)
			if err != nil {
				return apps.NewErrorResponse(err)
			}
			m, err = a.proxy.UpdateAppListing(appclient.UpdateAppListingRequest{
				Manifest:       *m,
				AddDeployments: apps.DeployTypes{apps.DeployHTTP},
			})
			if err != nil {
				return apps.NewErrorResponse(err)
			}

			return apps.NewFormResponse(*a.newInstallConsentForm(*m, creq, apps.DeployHTTP))
		},
	}
}
