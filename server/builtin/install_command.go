// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/pkg/errors"
)

const (
	DebugInstallFromURL = true
)

func (a *builtinApp) installCommandBinding() apps.Binding {
	installCommand := apps.Binding{
		Label:       "install",
		Location:    "install",
		Hint:        "[app source, e.g. marketplace]",
		Description: "Installs an App",
	}

	conf := a.conf.GetConfig()
	if conf.MattermostCloudMode {
		installCommand.Bindings = []apps.Binding{
			{
				Label:       "marketplace",
				Location:    "marketplace",
				Hint:        "[app ID]",
				Description: "Installs an App from the Marketplace",
				Call: &apps.Call{
					Path: pInstallS3,
				},
			},
		}
	} else {
		installCommand.Bindings = []apps.Binding{
			{
				Label:       "s3",
				Location:    "s3",
				Hint:        "[app ID]",
				Description: "Installs an App from AWS S3, as configured by the system administrator",
				Call: &apps.Call{
					Path: pInstallS3,
				},
			},
			{
				Label:       "url",
				Location:    "url",
				Hint:        "[manifest.json URL]",
				Description: "Installs an App from an HTTP URL",
				Call: &apps.Call{
					Path: pInstallURL,
				},
			},
		}
	}
	return installCommand
}

func (a *builtinApp) installS3Form(creq apps.CallRequest) apps.CallResponse {
	return formResponse(apps.Form{
		Title: "Install an App from AWS S3",
		Fields: []apps.Field{
			{
				Name:                 fAppID,
				Type:                 apps.FieldTypeDynamicSelect,
				Description:          "select an App",
				Label:                fAppID,
				AutocompleteHint:     "App ID",
				AutocompletePosition: 1,
			},
		},
		Call: &apps.Call{
			Path: pInstallS3,
		},
	})
}

func (a *builtinApp) installURLForm(creq apps.CallRequest) apps.CallResponse {
	return formResponse(apps.Form{
		Title: "Install an App from an HTTP URL",
		Fields: []apps.Field{
			{
				Name:                 fURL,
				Type:                 apps.FieldTypeText,
				Description:          "URL of the App manifest",
				Label:                fURL,
				AutocompleteHint:     "enter the URL",
				AutocompletePosition: 1,
			},
		},
		Call: &apps.Call{
			Path: pInstallURL,
		},
	})
}

func (a *builtinApp) installLookup(creq apps.CallRequest) apps.CallResponse {
	name := creq.GetValue("name", "")
	input := creq.GetValue("user_input", "")

	switch name {
	case fAppID:
		marketplaceApps := a.proxy.GetListedApps(input, false)
		var options []*apps.SelectOption
		for _, mapp := range marketplaceApps {
			if !mapp.Installed {
				options = append(options, &apps.SelectOption{
					Value: string(mapp.Manifest.AppID),
					Label: mapp.Manifest.DisplayName,
				})
			}
		}
		return dataResponse(options)
	}
	return apps.NewErrorCallResponse(errors.Errorf("unknown field %s", name))
}

func (a *builtinApp) installS3Submit(creq apps.CallRequest) apps.CallResponse {
	appID := apps.AppID(creq.GetValue(fAppID, ""))
	m, err := a.store.Manifest.Get(appID)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}

	return formResponse(
		a.newInstallConsentForm(*m, creq))
}

func (a *builtinApp) installURLSubmit(creq apps.CallRequest) apps.CallResponse {
	manifestURL := creq.GetValue(fURL, "")
	conf := a.conf.GetConfig()
	data, err := a.httpOut.GetFromURL(manifestURL, conf.DeveloperMode)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	m, err := apps.ManifestFromJSON(data)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}

	return formResponse(
		a.newInstallConsentForm(*m, creq))
}
