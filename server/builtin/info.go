// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

var infoCommandBinding = apps.Binding{
	Label:       "info",
	Location:    "info",
	Description: "Display Apps plugin info",
	Form:        blankForm(apps.NewCall(pInfo)),
}

func (a *builtinApp) info(creq apps.CallRequest) apps.CallResponse {
	conf := a.conf.Get()
	return apps.NewTextResponse("Mattermost Apps plugin version: %s, "+
		"[%s](https://github.com/mattermost/%s/commit/%s), built %s, Cloud Mode: %t, Developer Mode: %t\n",
		conf.Version,
		conf.BuildHashShort,
		config.Repository,
		conf.BuildHash,
		conf.BuildDate,
		conf.MattermostCloudMode,
		conf.DeveloperMode,
	)
}
