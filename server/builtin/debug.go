// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func (a *builtinApp) debugCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.label",
			Other: "debug",
		}),
		Location: "debug",
		Bindings: []apps.Binding{
			a.debugBindings().commandBinding(loc),
			a.debugClean().commandBinding(loc),
		},
	}
}

func (a *builtinApp) debugBindings() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func(loc *i18n.Localizer) apps.Binding {
			return apps.Binding{
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.bindings.label",
					Other: "bindings",
				}),
				Location: "bindings",
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.bindings.description",
					Other: "Display all bindings for the current context",
				}),
				Call: &apps.Call{
					Path: pDebugBindings,
					Expand: &apps.Expand{
						AdminAccessToken: apps.ExpandAll, // ensure sysadmin
					},
				},
				Form: &noParameters,
			}
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

func (a *builtinApp) debugClean() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func(loc *i18n.Localizer) apps.Binding {
			return apps.Binding{
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.clean.label",
					Other: "clean",
				}),
				Location: "clean",
				Hint:     "",
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.clean.description",
					Other: "remove all Apps and reset the persistent store",
				}),
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
			loc := i18n.NewLocalizer(a.conf.I18N().Bundle, creq.Context.Locale)
			_ = a.conf.MattermostAPI().KV.DeleteAll()
			_ = a.conf.StoreConfig(config.StoredConfig{})
			return apps.NewTextResponse(a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "command.debug.clean.submit.ok",
				Other: "Deleted all KV records and emptied the config.",
			}))
		},
	}
}
