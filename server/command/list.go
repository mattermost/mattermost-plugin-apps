// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (s *service) executeList(params *params) (*model.CommandResponse, error) {
	listed := s.proxy.GetListedApps("")
	installed := s.proxy.GetInstalledApps()

	txt := md.MD("| Name | Status | Version | Account | Locations | Permissions |\n")
	txt += md.MD("| :-- |:-- | :-- | :-- | :-- | :-- |\n")

	for _, app := range installed {
		m, _ := s.proxy.GetManifest(app.AppID)
		if m == nil {
			continue
		}

		status := "**Installed**"
		if app.Disabled {
			status += ", **Disabled**"
		}
		status += fmt.Sprintf(", type: `%s`", app.AppType)

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

		name := fmt.Sprintf("**[%s](%s)** (`/%s`)",
			app.DisplayName, app.HomepageURL, app.AppID)

		txt += md.Markdownf("|%s|%s|%s|%s|%s|%s|\n",
			name, status, version, account, app.GrantedLocations, app.GrantedPermissions)
	}

	for _, l := range listed {
		app, _ := s.proxy.GetInstalledApp(l.Manifest.AppID)
		if app != nil {
			continue
		}

		version := string(l.Manifest.Version)
		status := fmt.Sprintf("type: `%s`", l.Manifest.AppType)

		name := fmt.Sprintf("[%s](%s) (`/%s`)",
			l.Manifest.DisplayName, l.Manifest.HomepageURL, l.Manifest.AppID)
		txt += md.Markdownf("|%s|%s|%s|%s|%s|%s|\n",
			name, status, version, "", l.Manifest.RequestedLocations, l.Manifest.RequestedPermissions)
	}

	return out(params, txt)
}
