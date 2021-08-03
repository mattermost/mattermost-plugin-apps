// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
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
				Call:        &installMarketplaceCall,
			},
		}
	} else {
		installCommand.Bindings = []apps.Binding{
			{
				Label:       "s3",
				Location:    "s3",
				Hint:        "[app ID]",
				Description: "Installs an App from AWS S3, as configured by the system administrator",
				Call:        &installS3Call,
			},
			{
				Label:       "url",
				Location:    "url",
				Hint:        "[manifest.json URL]",
				Description: "Installs an App from an HTTP URL",
				Call:        &installURLCall,
			},
		}
	}
	return installCommand
}

func (a *builtinApp) installCommandSubmit(m apps.Manifest, creq apps.CallRequest) apps.CallResponse {
	err := a.store.Manifest.StoreLocal(m)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	return formResponse(a.newInstallConsentForm(m, creq))
}
