// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (a *builtinApp) debugCommandBinding() apps.Binding {
	return apps.Binding{
		Label:    "debug",
		Location: "debug",
		Bindings: []apps.Binding{
			{
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
			},
			{
				Label:       "bindings",
				Location:    "bindings",
				Description: "Display all bindings for the current context",
				Call: &apps.Call{
					Path: pDebugBindings,
					Expand: &apps.Expand{
						AdminAccessToken: apps.ExpandAll, // ensure sysadmin
					},
				},
				Form: &noParameters,
			},
		},
	}
}

func (a *builtinApp) debugClean(creq apps.CallRequest) apps.CallResponse {
	_ = a.conf.MattermostAPI().KV.DeleteAll()
	_ = a.conf.StoreConfig(config.StoredConfig{})
	return mdResponse("Deleted all KV records and emptied the config.")
}

func (a *builtinApp) debugBindings(creq apps.CallRequest) apps.CallResponse {
	bindings, err := a.proxy.GetBindings(proxy.NewIncomingFromContext(creq.Context), creq.Context)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	return mdResponse(utils.JSONBlock(bindings))
}
