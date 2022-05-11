// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (a *builtinApp) listCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.list.label",
			Other: "list",
		}),
		Location: "list",
		Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.list.hint",
			Other: "[ flags ]",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.list.description",
			Other: "Display available and installed Apps",
		}),
		Form: &apps.Form{
			Title: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "command.list.form.title",
				Other: "list Apps",
			}),
			Fields: []apps.Field{
				{
					Name: fIncludePlugins,
					Type: apps.FieldTypeBool,
					Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "field.include_plugins.label",
						Other: "include-plugins",
					}),
					Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "field.include_plugins.description",
						Other: "include compatible Mattermost plugins in the output.",
					}),
				},
			},
			Submit: newUserCall(pList).WithLocale(),
		},
	}
}

func (a *builtinApp) list(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	loc := a.newLocalizer(creq)
	includePluginApps := creq.BoolValue("plugin-apps")

	listed := a.proxy.GetListedApps("", includePluginApps)
	installed, reachable := a.proxy.GetInstalledApps(r, true)

	// All of this information is non sensitive.
	// Checks for the user's permissions might be needed in the future.
	txt := a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
		ID:    "command.list.submit.header",
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
			ID:    "command.list.submit.status.installed",
			Other: "**Installed**",
		})
		if app.Disabled {
			status = a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "command.list.submit.status.disabled",
				Other: "Installed, Disabled",
			})
		} else {
			if !reachable[app.AppID] {
				status = a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.list.submit.status.unreachable",
					Other: "Installed, **Unreachable**",
				})
			}
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
		if app.MattermostOAuth2 != nil {
			if account != "" {
				account += ", "
			}
			account += fmt.Sprintf("OAuth: `%s`", app.MattermostOAuth2.Id)
			if app.RemoteOAuth2.ClientID != "" {
				account += fmt.Sprintf("/`%s`", app.RemoteOAuth2.ClientID)
			}
		}

		name := fmt.Sprintf("**[%s](%s)** (`%s`)",
			app.DisplayName, app.HomepageURL, app.AppID)

		deployType := string(app.DeployType)
		if app.DeployType == apps.DeployHTTP && app.HTTP != nil {
			deployType = app.HTTP.RootURL
		}

		txt += fmt.Sprintf("|%s|%s|%s|%s|%s|%s|%s|\n",
			name, status, deployType, version, account, app.GrantedLocations, app.GrantedPermissions)
	}

	listedString := a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
		ID:    "command.list.submit.listed",
		Other: "Listed",
	})

	for _, l := range listed {
		app, _ := a.proxy.GetInstalledApp(l.Manifest.AppID, false)
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
