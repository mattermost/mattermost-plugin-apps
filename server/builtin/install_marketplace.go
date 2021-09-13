// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var installMarketplaceCall = apps.Call{
	Path: pInstallMarketplace,
	Expand: &apps.Expand{
		AdminAccessToken: apps.ExpandAll, // ensure sysadmin
	},
}

func (a *builtinApp) installMarketplace() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func() apps.Binding {
			return apps.Binding{
				Label:       "marketplace",
				Location:    "marketplace",
				Hint:        "[app ID]",
				Description: "Installs an App from the Marketplace",
				Call:        &installMarketplaceCall,
				Form:        appIDForm(installMarketplaceCall),
			}
		},

		lookupf: func(creq apps.CallRequest) ([]apps.SelectOption, error) {
			return a.lookupAppID(creq, func(app apps.ListedApp) bool {
				return !app.Installed
			})
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			appID := apps.AppID(creq.GetValue(fAppID, ""))
			m, err := a.store.Manifest.Get(appID)
			if err != nil {
				return apps.NewErrorCallResponse(err)
			}

			return a.installCommandSubmit(*m, creq)
		},
	}
}
