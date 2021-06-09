// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/utils/md"
)

func (s *service) executeList(params *commandParams) (*model.CommandResponse, error) {
	listed := s.proxy.GetListedApps("")
	installed := s.proxy.GetInstalledApps()

	txt := md.MD("| Name | Status | Type | Version | Account | Locations | Permissions |\n")
	txt += md.MD("| :-- |:-- | :-- | :-- | :-- | :-- | :-- |\n")

	for _, app := range installed {
		m, _ := s.proxy.GetManifest(app.AppID)
		if m == nil {
			continue
		}

		status := "**Installed**"
		if app.Disabled {
			status += ", **Disabled**"
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

		txt += md.Markdownf("|%s|%s|%s|%s|%s|%s|%s|\n",
			name, status, app.AppType, version, account, app.GrantedLocations, app.GrantedPermissions)
	}

	for _, l := range listed {
		app, _ := s.proxy.GetInstalledApp(l.Manifest.AppID)
		if app != nil {
			continue
		}

		status := "Listed"

		version := string(l.Manifest.Version)

		name := fmt.Sprintf("[%s](%s) (`%s`)",
			l.Manifest.DisplayName, l.Manifest.HomepageURL, l.Manifest.AppID)
		txt += md.Markdownf("|%s|%s|%s|%s|%s|%s|%s|\n",
			name, status, l.Manifest.AppType, version, "", l.Manifest.RequestedLocations, l.Manifest.RequestedPermissions)
	}

	return out(params, txt)
}
