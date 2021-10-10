// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (a *builtinApp) installCommandBinding() apps.Binding {
	if a.conf.Get().MattermostCloudMode {
		return apps.Binding{
			Label:       "install",
			Location:    "install",
			Hint:        "[app ID]",
			Description: "Installs an App from the Marketplace",
			Form:        appIDForm(newAdminCall(pInstallListed), newAdminCall(pInstallListedLookup)),
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
					Form:        appIDForm(newAdminCall(pInstallListed), newAdminCall(pInstallListedLookup)),
				},
				{
					Label:       "url",
					Location:    "url",
					Hint:        "[manifest.json URL]",
					Description: "Installs an App from an HTTP URL",
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
						Submit: newAdminCall(pInstallURL),
					},
				},
			},
		}
	}
}

func (a *builtinApp) installListedLookup(creq apps.CallRequest) apps.CallResponse {
	return a.lookupAppID(creq, func(app apps.ListedApp) bool {
		return !app.Installed
	})
}

func (a *builtinApp) installListed(creq apps.CallRequest) apps.CallResponse {
	appID := apps.AppID(creq.GetValue(fAppID, ""))
	m, err := a.proxy.GetManifest(appID)
	if err != nil {
		return apps.NewErrorResponse(err)
	}

	return apps.NewFormResponse(a.newInstallConsentForm(*m, creq))
}

func (a *builtinApp) installURL(creq apps.CallRequest) apps.CallResponse {
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
	_, err = a.proxy.StoreLocalManifest(*m)
	if err != nil {
		return apps.NewErrorResponse(err)
	}

	return apps.NewFormResponse(a.newInstallConsentForm(*m, creq))
}
