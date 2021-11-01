// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func (a *builtinApp) list() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func(loc *i18n.Localizer) apps.Binding {
			return apps.Binding{
				Location:    "list",
				Label:       a.conf.Local(loc, "command.list.label"),
				Description: a.conf.Local(loc, "command.list.description"),
				Hint:        a.conf.Local(loc, "command.list.hint"),
				Call: &apps.Call{
					Path: pList,
					Expand: &apps.Expand{
						ActingUser: apps.ExpandSummary,
						Locale:     apps.ExpandAll,
					},
				},
				Form: &apps.Form{
					Fields: []apps.Field{
						{
							Name:        fIncludePlugins,
							Type:        apps.FieldTypeBool,
							Label:       a.conf.Local(loc, "field.include_plugins.label"),
							Description: a.conf.Local(loc, "field.include_plugins.description"),
						},
					},
				},
			}
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			loc := a.newLocalizer(creq)
			includePluginApps := creq.BoolValue(fIncludePlugins)

			listed := a.proxy.GetListedApps("", includePluginApps)
			installed := a.proxy.GetInstalledApps()

			// All of this information is non sensitive.
			// Checks for the user's permissions might be needed in the future.
			txt := a.conf.Local(loc, "command.list.submit.header")
			txt += "\n| :-- |:-- | :-- | :-- | :-- | :-- | :-- |\n"

			for _, app := range installed {
				m, _ := a.proxy.GetManifest(app.AppID)
				if m == nil {
					continue
				}

				if !includePluginApps && app.DeployType == apps.DeployPlugin {
					continue
				}

				status := a.conf.Local(loc, "command.list.submit.status.installed")
				if app.Disabled {
					status = a.conf.Local(loc, "command.list.submit.status.disabled")
				}

				version := string(app.Version)
				if string(m.Version) != version {
					version = a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
						DefaultMessage: &i18n.Message{
							ID:    "command.list.submit.version",
							Other: "{{.CurrentVersion}}, {{.MarketplaceVersion}} in marketplace",
						},
						TemplateData: map[string]string{
							"CurrentVersion":     string(app.Version),
							"MarketplaceVersion": string(m.Version),
						},
					})
				}

				// TODO Translate the account part
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
					name, status, app.DeployType, version, account, app.GrantedLocations, app.GrantedPermissions)
			}

			listedString := a.conf.Local(loc, "command.list.submit.listed")
			for _, l := range listed {
				app, _ := a.proxy.GetInstalledApp(l.Manifest.AppID)
				if app != nil {
					continue
				}
				status := listedString
				version := string(l.Manifest.Version)
				name := fmt.Sprintf("[%s](%s) (`%s`)",
					l.Manifest.DisplayName, l.Manifest.HomepageURL, l.Manifest.AppID)

				txt += fmt.Sprintf("|%s|%s|%s|%s|%s|%s|%s|\n",
					name, status, l.Manifest.DeployTypes(), version, "", l.Manifest.RequestedLocations, l.Manifest.RequestedPermissions)
			}
			return apps.NewTextResponse(txt)
		},
	}
}
