// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/hashicorp/go-getter"
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

const (
	DebugInstallFromURL = true
)

func (a *builtinApp) installCommandBinding() *apps.Binding {
	installCommand := &apps.Binding{
		Label:       "install",
		Location:    "install",
		Hint:        "[app source, e.g. marketplace]",
		Description: "Installs an App",
	}

	conf := a.conf.GetConfig()
	if conf.MattermostCloudMode {
		installCommand.Bindings = []*apps.Binding{
			{
				Label:       "marketplace",
				Location:    "marketplace",
				Hint:        "[app ID]",
				Description: "Installs an App from the Marketplace",
				Call: &apps.Call{
					Path: pInstallMarketplace,
				},
			},
		}
	} else {
		installCommand.Bindings = []*apps.Binding{
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

func (a *builtinApp) installMarketplaceForm(creq *apps.CallRequest) *apps.CallResponse {
	return responseForm(&apps.Form{
		Title: "Install an App from marketplace",
		Fields: []*apps.Field{
			{
				Name:                 fAppID,
				Type:                 apps.FieldTypeDynamicSelect,
				Description:          "select a Marketplace App",
				Label:                fAppID,
				AutocompleteHint:     "App ID",
				AutocompletePosition: 1,
			},
		},
		Call: &apps.Call{
			Path: pInstallMarketplace,
		},
	})
}

func (a *builtinApp) installS3Form(creq *apps.CallRequest) *apps.CallResponse {
	return responseForm(&apps.Form{
		Title: "Install an App from AWS S3",
		Fields: []*apps.Field{
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

func (a *builtinApp) installHTTPForm(creq *apps.CallRequest) *apps.CallResponse {
	return responseForm(&apps.Form{
		Title: "Install an App from an HTTP URL",
		Fields: []*apps.Field{
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

func (a *builtinApp) installMarketplaceLookup(creq *apps.CallRequest) *apps.CallResponse {
	name := creq.GetStringValue("name", "")
	input := creq.GetStringValue("user_input", "")

	switch name {
	case fAppID:
		marketplaceApps := a.proxy.ListMarketplaceApps(input)
		var options []*apps.SelectOption
		for _, mapp := range marketplaceApps {
			if !mapp.Installed {
				options = append(options, &apps.SelectOption{
					Value: string(mapp.Manifest.AppID),
					Label: mapp.Manifest.DisplayName,
				})
			}
		}
		return options
	}
	return nil
}

func (a *builtinApp) installMarketplaceSubmit(creq *apps.CallRequest) *apps.CallResponse {
	appID := apps.AppID(creq.GetValue(fAppID, ""))
	m, err := a.store.Manifest.Get(appID)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}

	return a.installAppFormManifest(m, creq)
}

func (a *builtinApp) installURLSubmit(creq *apps.CallRequest) *apps.CallResponse {
	url := creq.GetValue(fURL, "")

	data, err := getter.Get(url)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	m, err := apps.ManifestFromJSON(data)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	return a.installAppFormManifest(m, creq)
}
