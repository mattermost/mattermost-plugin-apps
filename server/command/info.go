// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

func (s *service) executeInfo(params *commandParams) *model.CommandResponse {
	loc := s.conf.I18N().GetUserLocalizer(params.commandArgs.UserId)

	conf := s.conf.Get()
	resp := s.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "apps.command.info.version",
			Other: "Mattermost Apps plugin version: {{.Version}}, ",
		},
		TemplateData: map[string]string{
			"Version": conf.Version,
		},
	}) + s.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "apps.command.info.rest",
			Other: "{{.URL}}, built {{.BuildDate}}, Cloud Mode: {{.CloudMode}}, Developer Mode: {{.DeveloperMode}}",
		},
		TemplateData: map[string]string{
			"URL":           fmt.Sprintf("[%s](https://github.com/mattermost/%s/commit/%s)", conf.BuildHashShort, config.Repository, conf.BuildHash),
			"BuildDate":     conf.BuildDate,
			"CloudMode":     fmt.Sprintf("%t", conf.MattermostCloudMode),
			"DeveloperMode": fmt.Sprintf("%t", conf.DeveloperMode),
		},
	}) + "\n"

	return out(params, resp)
}
