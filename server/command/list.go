// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (s *service) executeList(params *commandParams) (*model.CommandResponse, error) {
	loc := s.conf.I18N().GetUserLocalizer(params.commandArgs.UserId)

	var includePluginApps bool
	fs := pflag.NewFlagSet("plugin-apps", pflag.ContinueOnError)
	fs.BoolVar(&includePluginApps, "plugin-apps", false, "Include apps managed by plugins")
	err := fs.Parse(params.current)
	if err != nil {
		return s.errorOut(params, err)
	}

	listed := s.proxy.GetListedApps("", includePluginApps)
	installed := s.proxy.GetInstalledApps()

	txt := s.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
		ID:    "apps.command.list.table.header",
		Other: "| Name | Status | Type | Version | Account | Locations | Permissions |",
	}) + "\n"
	txt += "| :-- |:-- | :-- | :-- | :-- | :-- | :-- |\n"

	for _, app := range installed {
		m, _ := s.proxy.GetManifest(app.AppID)
		if m == nil {
			continue
		}

		if !includePluginApps && m.AppType == apps.AppTypePlugin {
			continue
		}

		status := s.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "apps.command.list.status.installed",
			Other: "**Installed**",
		})
		if app.Disabled {
			status = s.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "apps.command.list.status.disabled",
				Other: "Installed, Disabled",
			})
		}

		version := string(app.Version)
		if string(m.Version) != version {
			version += s.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "apps.command.list.version.marketplace",
					Other: ", {{.Version}} in marketplace",
				},
				TemplateData: map[string]string{
					"Version": string(m.Version),
				},
			})
		}

		account := ""
		if app.BotUserID != "" {
			account += fmt.Sprintf("Bot: `%s`", app.BotUserID)
		}
		if app.MattermostOAuth2.ClientID != "" {
			if account != "" {
				account += ", "
			}
			account += fmt.Sprintf("OAuth: `%s`", app.MattermostOAuth2.ClientID)
			if app.RemoteOAuth2.ClientID != "" {
				account += fmt.Sprintf("/`%s`", app.RemoteOAuth2.ClientID)
			}
		}

		name := fmt.Sprintf("**[%s](%s)** (`%s`)",
			app.DisplayName, app.HomepageURL, app.AppID)

		txt += fmt.Sprintf("|%s|%s|%s|%s|%s|%s|%s|\n",
			name, status, app.AppType, version, account, app.GrantedLocations, app.GrantedPermissions)
	}

	for _, l := range listed {
		app, _ := s.proxy.GetInstalledApp(l.Manifest.AppID)
		if app != nil {
			continue
		}

		status := s.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "apps.command.list.listed",
			Other: "Listed",
		})

		version := string(l.Manifest.Version)

		name := fmt.Sprintf("[%s](%s) (`%s`)",
			l.Manifest.DisplayName, l.Manifest.HomepageURL, l.Manifest.AppID)
		txt += fmt.Sprintf("|%s|%s|%s|%s|%s|%s|%s|\n",
			name, status, l.Manifest.AppType, version, "", l.Manifest.RequestedLocations, l.Manifest.RequestedPermissions)
	}

	return out(params, txt)
}
