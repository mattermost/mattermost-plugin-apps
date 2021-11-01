// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var debugBindingsCall = apps.Call{
	Path: pDebugBindings,
	Expand: &apps.Expand{
		ActingUser:       apps.ExpandSummary,
		AdminAccessToken: apps.ExpandAll,
	},
}

func (a *builtinApp) debugBindings() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func(loc *i18n.Localizer) apps.Binding {
			return apps.Binding{
				Location:    "bindings",
				Label:       a.conf.Local(loc, "command.debug.bindings.label"),
				Description: a.conf.Local(loc, "command.debug.bindings.description"),
				Call:        &debugBindingsCall,
				Form:        &noParameters,
			}
		},

		lookupf: func(creq apps.CallRequest) ([]apps.SelectOption, error) {
			return a.lookupAppID(creq, func(app apps.ListedApp) bool {
				return app.Installed && app.Enabled
			})
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			bindings, err := a.proxy.GetBindings(proxy.NewIncomingFromContext(creq.Context), creq.Context)
			if err != nil {
				return apps.NewErrorResponse(err)
			}
			return apps.NewTextResponse(utils.JSONBlock(bindings))

		},
	}
}
