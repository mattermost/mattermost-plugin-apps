// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (a *builtinApp) listCommandBinding() apps.Binding {
	return apps.Binding{
		Label:       "list",
		Location:    "list",
		Hint:        "[ flags ]",
		Description: "Display available and installed Apps",
		Call: &apps.Call{
			Path: pList,
		},
		Form: &apps.Form{
			Title: "list Apps",
			Fields: []apps.Field{
				{
					Label: "include-plugins",
					Name:  fIncludePlugins,
					Type:  apps.FieldTypeBool,
				},
			},
		},
	}
}

func (a *builtinApp) list(creq apps.CallRequest) apps.CallResponse {
	includePluginApps := creq.BoolValue("plugin-apps")

	listed := a.proxy.GetListedApps("", includePluginApps)
	installed := a.proxy.GetInstalledApps()

	// All of this information is non sensitive.
	// Checks for the user's permissions might be needed in the future.
	txt := "| Name | Status | Type | Version | Account | Locations | Permissions |\n"
	txt += "| :-- |:-- | :-- | :-- | :-- | :-- | :-- |\n"

	for _, app := range installed {
		m, _ := a.proxy.GetManifest(app.AppID)
		if m == nil {
			continue
		}

		if !includePluginApps && app.AppType == apps.AppTypePlugin {
			continue
		}

		status := "**Installed**"
		if app.Disabled {
			status = "Installed, Disabled"
		}

		version := string(app.Version)
		if string(m.Version) != version {
			version += fmt.Sprintf(", %s in marketplace", m.Version)
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
		app, _ := a.proxy.GetInstalledApp(l.Manifest.AppID)
		if app != nil {
			continue
		}

		status := "Listed"

		version := string(l.Manifest.Version)

		name := fmt.Sprintf("[%s](%s) (`%s`)",
			l.Manifest.DisplayName, l.Manifest.HomepageURL, l.Manifest.AppID)
		txt += fmt.Sprintf("|%s|%s|%s|%s|%s|%s|%s|\n",
			name, status, l.Manifest.AppType, version, "", l.Manifest.RequestedLocations, l.Manifest.RequestedPermissions)
	}
	return mdResponse(txt)
}
