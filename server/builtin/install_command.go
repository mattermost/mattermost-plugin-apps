// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
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
			{
				Name:                 fVersion,
				Type:                 apps.FieldTypeDynamicSelect,
				Description:          "select the App's version",
				Label:                fVersion,
				AutocompleteHint:     "app version",
				AutocompletePosition: 2,
			},
		},
		Call: &apps.Call{
			Path: pInstallS3,
		},
	})
}

type lookupResponse struct {
	Items []apps.SelectOption `json:"items"`
}

func (a *builtinApp) installS3Lookup(creq apps.CallRequest) apps.CallResponse {
	if creq.SelectedField != fAppID && creq.SelectedField != fVersion {
		return apps.NewErrorCallResponse(errors.Errorf("unknown field %q", creq.SelectedField))
	}

	conf := a.conf.GetConfig()
	up, err := upaws.MakeUpstream(conf.AWSAccessKey, conf.AWSSecretKey, conf.AWSRegion, conf.AWSS3Bucket, a.log)
	if err != nil {
		return apps.NewErrorCallResponse(errors.Wrap(err, "failed to initialize AWS access"))
	}

	var options []apps.SelectOption
	switch creq.SelectedField {
	case fAppID:
		appIDs, err := up.ListS3Apps(creq.Query)
		if err != nil {
			return apps.NewErrorCallResponse(errors.Wrap(err, "failed to retrive the list of apps, try --url"))
		}
		for _, appID := range appIDs {
			options = append(options, apps.SelectOption{
				Value: string(appID),
				Label: string(appID),
			})
		}

	case fVersion:
		id := creq.GetValue(fAppID, "")
		versions, err := up.ListS3Versions(apps.AppID(id), creq.Query)
		if err != nil {
			return apps.NewErrorCallResponse(errors.Wrap(err, "failed to retrive the list of apps, try --url"))
		}
		for _, v := range versions {
			options = append(options, apps.SelectOption{
				Value: string(v),
				Label: string(v),
			})
		}
	}

	return dataResponse(
		lookupResponse{
			Items: options,
		})
}

func (a *builtinApp) installS3Submit(creq apps.CallRequest) apps.CallResponse {
	appID := apps.AppID(creq.GetValue(fAppID, ""))
	version := apps.AppVersion(creq.GetValue(fVersion, ""))
	m, err := a.store.Manifest.GetFromS3(appID, version)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}

	return a.installCommandSubmit(*m, creq)
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

	return a.installCommandSubmit(*m, creq)
}

func (a *builtinApp) installMarketplaceForm(creq apps.CallRequest) apps.CallResponse {
	return formResponse(apps.Form{
		Title: "Install an App from Mattermost Apps Marketplace",
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
			Path: pInstallMarketplace,
		},
	})
}

func (a *builtinApp) installMarketplaceLookup(creq apps.CallRequest) apps.CallResponse {
	if creq.SelectedField != fAppID {
		return apps.NewErrorCallResponse(errors.Errorf("unknown field %q", creq.SelectedField))
	}

	var options []apps.SelectOption
	marketplaceApps := a.proxy.GetListedApps(creq.Query, false)
	for _, app := range marketplaceApps {
		options = append(options, apps.SelectOption{
			Value: string(app.Manifest.AppID),
			Label: string(app.Manifest.DisplayName),
		})
	}

	return dataResponse(
		lookupResponse{
			Items: options,
		})
}

func (a *builtinApp) installMarketplaceSubmit(creq apps.CallRequest) apps.CallResponse {
	appID := apps.AppID(creq.GetValue(fAppID, ""))
	m, err := a.store.Manifest.Get(appID)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}

	return a.installCommandSubmit(*m, creq)
}

func (a *builtinApp) installCommandSubmit(m apps.Manifest, creq apps.CallRequest) apps.CallResponse {
	err := a.store.Manifest.StoreLocal(m)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	return formResponse(a.newInstallConsentForm(m, creq))
}
