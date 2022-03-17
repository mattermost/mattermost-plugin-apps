// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
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
			ID:    "command.info.submit",
			Other: "Mattermost Apps plugin version: {{.Version}}, {{.URL}}, built {{.BuildDate}}, Cloud Mode: {{.CloudMode}}, Developer Mode: {{.DeveloperMode}}, Allow HTTP Apps: {{.AllowHTTPApps}}",
		},
		TemplateData: conf.InfoTemplateData(),
	}) + "\n"
	return apps.NewTextResponse(out)
}
