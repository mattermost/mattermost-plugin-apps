// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
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

func (a *builtinApp) info(_ *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	loc := a.newLocalizer(creq)
	conf := a.conf.Get()
	out := a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "command.info.submit",
			Other: "Mattermost Apps plugin version: {{.Version}}, {{.URL}}, built {{.BuildDate}}\n\n- Mattermost Cloud Mode: {{.CloudMode}}\n- Developer Mode: {{.DeveloperMode}}\n- Allow HTTP Apps: {{.AllowHTTPApps}}",
		},
		TemplateData: conf.InfoTemplateData(),
	}) + "\n\n"

	if conf.DeveloperMode && conf.AWSAccessKey != "" {
		out += a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "command.info.aws",
				Other: "AWS config:\n- Region: `{{.Region}}`\n- S3 Bucket: `{{.Bucket}}`\n- Access Key: `{{.Access}}`\n- Secret Key: `{{.Secret}}`",
			},
			TemplateData: map[string]string{
				"Region": conf.AWSRegion,
				"Bucket": conf.AWSS3Bucket,
				"Access": utils.LastN(conf.AWSAccessKey, 4),
				"Secret": utils.LastN(conf.AWSSecretKey, 4),
			},
		}) + "\n"
	}
	return apps.NewTextResponse(out)
}
