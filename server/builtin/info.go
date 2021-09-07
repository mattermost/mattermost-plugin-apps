// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

func (a *builtinApp) infoCommandBinding() apps.Binding {
	return apps.Binding{
		Label:       "info",
		Location:    "info",
		Description: "Display Apps plugin info",
		Call: &apps.Call{
			Path: pInfo,
		},
		Form: &noParameters,
	}
}

func (a *builtinApp) info(creq apps.CallRequest) apps.CallResponse {
	conf := a.conf.Get()
	return mdResponse("Mattermost Cloud Apps plugin version: %s, "+
		"[%s](https://github.com/mattermost/%s/commit/%s), built %s\n",
		conf.Version,
		conf.BuildHashShort,
		config.Repository,
		conf.BuildHash,
		conf.BuildDate)
}
