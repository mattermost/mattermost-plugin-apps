// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var installMarketplaceCall = apps.Call{
	Path: pInstallMarketplace,
}

func (a *builtinApp) installMarketplaceForm(creq apps.CallRequest) apps.CallResponse {
	return appIDForm(installMarketplaceCall)
}

func (a *builtinApp) installMarketplaceLookup(creq apps.CallRequest) apps.CallResponse {
	return a.lookupAppID(creq, func(app apps.ListedApp) bool {
		return !app.Installed
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
