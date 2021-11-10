// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

func (a *builtinApp) infoCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.info.label",
			Other: "info",
		}),
		Location: "info",
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.info.description",
			Other: "Display Apps plugin info",
		}),
		Submit: &apps.Call{
			Path: pInfo,
			Expand: &apps.Expand{
				Locale: apps.ExpandAll,
			},
		},
	}
}

func (a *builtinApp) info(creq apps.CallRequest) apps.CallResponse {
	loc := i18n.NewLocalizer(a.conf.I18N().Bundle, creq.Context.Locale)
	conf := a.conf.Get()
	out := a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "apps.command.info.submit.ok",
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
}
