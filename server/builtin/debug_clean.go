// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

func (a *builtinApp) debugClean() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func(loc *i18n.Localizer) apps.Binding {
			return apps.Binding{
				Location:    "clean",
				Label:       a.conf.Local(loc, "command.debug.clean.label"),
				Description: a.conf.Local(loc, "command.debug.clean.description"),
				Call: &apps.Call{
					Path: pDebugClean,
					Expand: &apps.Expand{
						AdminAccessToken: apps.ExpandAll, // ensure sysadmin
						Locale:           apps.ExpandAll,
					},
				},
				Form: &noParameters,
			}
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			loc := a.newLocalizer(creq)
			_ = a.conf.MattermostAPI().KV.DeleteAll()
			_ = a.conf.StoreConfig(config.StoredConfig{})
			return apps.NewTextResponse(a.conf.Local(loc, "command.debug.clean.submit"))
		},
	}
}
