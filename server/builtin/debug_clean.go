// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

func (a *builtinApp) debugClean() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func() apps.Binding {
			return apps.Binding{
				Label:       "clean",
				Location:    "clean",
				Hint:        "",
				Description: "remove all Apps and reset the persistent store",
				Call: &apps.Call{
					Path: pDebugClean,
					Expand: &apps.Expand{
						AdminAccessToken: apps.ExpandAll, // ensure sysadmin
					},
				},
				Form: &noParameters,
			}
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			_ = a.conf.MattermostAPI().KV.DeleteAll()
			_ = a.conf.StoreConfig(config.StoredConfig{})
			return apps.NewTextResponse("Deleted all KV records and emptied the config.")
		},
	}
}
