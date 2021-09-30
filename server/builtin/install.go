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

	conf := a.conf.Get()
	if conf.MattermostCloudMode {
		installCommand.Bindings = []apps.Binding{
			a.installMarketplace().commandBinding(),
		}
	} else {
		installCommand.Bindings = []apps.Binding{
			a.installS3().commandBinding(),
			a.installURL().commandBinding(),
		}
	}
	return installCommand
}

func (a *builtinApp) installCommandSubmit(m apps.Manifest, creq apps.CallRequest) apps.CallResponse {
	err := a.store.Manifest.StoreLocal(m)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	return formResponse(*a.newInstallConsentForm(m, creq))
}
