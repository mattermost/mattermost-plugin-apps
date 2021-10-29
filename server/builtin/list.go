// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func (a *builtinApp) listCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.enable.list.label",
			Other: "list",
		}),
		Location: "list",
		Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.enable.list.hint",
			Other: "[ flags ]",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.enable.list.description",
			Other: "Display available and installed Apps",
		}),
		Form: &apps.Form{
			Title: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "command.enable.list.form.title",
				Other: "list Apps",
			}),
			Fields: []apps.Field{
				{
					Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "command.enable.list.form.include_plugins.label",
						Other: "include-plugins",
					}),
					Name: fIncludePlugins,
					Type: apps.FieldTypeBool,
				},
			},
			Submit: newAdminCall(pList).WithLocale(),
		},
	}
}

func (a *builtinApp) list(creq apps.CallRequest) apps.CallResponse {
	loc := i18n.NewLocalizer(a.conf.I18N().Bundle, creq.Context.Locale)
	includePluginApps := creq.BoolValue("plugin-apps")

	listed := a.proxy.GetListedApps("", includePluginApps)
	installed := a.proxy.GetInstalledApps()

	// All of this information is non sensitive.
	// Checks for the user's permissions might be needed in the future.
	txt := a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
		ID:    "apps.command.list.submit.ok.header",
		Other: "| Name | Status | Type | Version | Account | Locations | Permissions |",
	}) + "\n"
	txt += "| :-- |:-- | :-- | :-- | :-- | :-- | :-- |\n"

	for _, app := range installed {
		m, _ := a.proxy.GetManifest(app.AppID)
		if m == nil {
			continue
		}

		if !includePluginApps && app.DeployType == apps.DeployPlugin {
			continue
		}

		status := a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "apps.command.list.submit.ok.status.installed",
			Other: "**Installed**",
		})
		if app.Disabled {
			status = a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "apps.command.list.submit.ok.status.disabled",
				Other: "Installed, Disabled",
			})
		}

		version := string(app.Version)
		if string(m.Version) != version {
			version = a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "apps.command.list.submit.ok.version",
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

	listedString := a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
		ID:    "apps.command.list.submit.ok.listed",
		Other: "Listed",
	})

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
}
