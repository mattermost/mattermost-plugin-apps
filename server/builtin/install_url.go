// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var installURLCall = apps.Call{
	Path: pInstallURL,
}

func (a *builtinApp) installURLForm(creq apps.CallRequest) apps.CallResponse {
	return appIDForm(installURLCall)
}

func (a *builtinApp) installURLSubmit(creq apps.CallRequest) apps.CallResponse {
	manifestURL := creq.GetValue(fURL, "")
	conf := a.conf.Get()
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
