// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

var viewOAuthConfigCall = apps.Call{
	Path: pDebugOAuthConfigView,
	Expand: &apps.Expand{
		ActingUser: apps.ExpandSummary,
	},
}

func (a *builtinApp) debugOAuthConfigView() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func(loc *i18n.Localizer) apps.Binding {
			return apps.Binding{
				Location: "view",
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.oauth.config.view.label",
					Other: "view",
				}),
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.oauth.config.view.description",
					Other: "View the OAuth configuration of a app.",
				}),
				Call: &viewOAuthConfigCall,
				Form: a.appIDForm(viewOAuthConfigCall, loc),
			}
		},

		// Lookup returns the list of eligible Apps.
		lookupf: func(r *incoming.Request, creq apps.CallRequest) ([]apps.SelectOption, error) {
			return a.lookupAppID(r, creq, nil)
		},

		submitf: func(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
			appID := apps.AppID(creq.GetValue(fAppID, ""))
			r.SetAppID(appID)

			app, err := a.proxy.GetInstalledApp(r, appID)
			if err != nil {
				return apps.NewErrorResponse(err)
			}

			return apps.NewTextResponse(utils.JSONBlock(app.RemoteOAuth2))
		},
	}
}
