// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (a *builtinApp) info() handler {
	return handler{
		requireSysadmin: false,

		commandBinding: func(loc *i18n.Localizer) apps.Binding {
			return apps.Binding{
				Location: "info",
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.info.label",
					Other: "info",
				}),
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.info.description",
					Other: "Display Apps plugin info",
				}),
				Call: &apps.Call{
					Path: pInfo,
					Expand: &apps.Expand{
						Locale: apps.ExpandAll,
					},
				},
				Form: &noParameters,
			}
		},

		submitf: func(_ *incoming.Request, creq apps.CallRequest) apps.CallResponse {
			loc := a.newLocalizer(creq)
			conf := a.conf.Get()
			out := a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "command.info.submit",
					Other: "Mattermost Apps plugin version: {{.Version}}, {{.URL}}, built {{.BuildDate}}, Cloud Mode: {{.CloudMode}}, Developer Mode: {{.DeveloperMode}}",
				},
				TemplateData: map[string]string{
					"Version":       conf.PluginManifest.Version,
					"URL":           fmt.Sprintf("[%s](https://github.com/mattermost/%s/commit/%s)", conf.BuildHashShort, config.Repository, conf.BuildHash),
					"BuildDate":     conf.BuildDate,
					"CloudMode":     fmt.Sprintf("%t", conf.MattermostCloudMode),
					"DeveloperMode": fmt.Sprintf("%t", conf.DeveloperMode),
				},
			}) + "\n"
			return apps.NewTextResponse(out)
		},
	}
}
