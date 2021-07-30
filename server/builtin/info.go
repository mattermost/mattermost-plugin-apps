// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

func (a *builtinApp) infoCommandBinding() apps.Binding {
	return commandBinding("info", pInfo, "", "display Apps plugin info")
}

// func (a *builtinApp) infoForm(c *apps.Call) *apps.CallResponse {
// 	return &apps.CallResponse{
// 		Type: apps.CallResponseTypeForm,
// 		Form: &apps.Form{
// 			Title: "Apps proxy info",
// 			Call: &apps.Call{
// 				Path: PathInfo,
// 			},
// 		},
// 	}
// }

func (a *builtinApp) info(creq apps.CallRequest) apps.CallResponse {
	conf := a.conf.GetConfig()
	return mdResponse("Mattermost Cloud Apps plugin version: %s, "+
		"[%s](https://github.com/mattermost/%s/commit/%s), built %s\n",
		conf.Version,
		conf.BuildHashShort,
		config.Repository,
		conf.BuildHash,
		conf.BuildDate)
}
