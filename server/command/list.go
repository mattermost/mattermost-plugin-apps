// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (s *service) executeList(params *params) (*model.CommandResponse, error) {
	listed := s.admin.GetListedApps("")
	installed := s.admin.GetInstalledApps()

	txt := md.MD("| Name | Status | Version | Account | Locations | Permissions |\n")
	txt += md.MD("| :-- |:-- | :-- | :-- | :-- | :-- |\n")

	for _, app := range installed {
		l := listed[app.AppID]
		if l == nil {
			continue
		}

		status := "**Installed**"
		if app.Disabled {
			status += ", **Disabled**"
		}
		status += fmt.Sprintf(", type: `%s`", app.Type)

		version := string(app.Version)
		if string(l.Manifest.Version) != version {
			version += fmt.Sprintf(", %s in marketplace", l.Manifest.Version)
		}

		account := ""
		if app.BotUserID != "" {
			account += fmt.Sprintf("Bot: `%s`", app.BotUserID)
		}
		if app.OAuth2ClientID != "" {
			if account != "" {
				account += ", "
			}
			account += fmt.Sprintf("OAuth: `%s`", app.OAuth2ClientID)
		}
		name := fmt.Sprintf("**[%s](%s)** (`/%s`)",
			app.DisplayName, app.HomepageURL, app.AppID)

		txt += md.Markdownf("|%s|%s|%s|%s|%s|%s|\n",
			name, status, version, account, app.GrantedLocations, app.GrantedPermissions)
	}

	for _, l := range listed {
		_, ok := installed[l.Manifest.AppID]
		if ok {
			continue
		}

		version := string(l.Manifest.Version)
		status := fmt.Sprintf("type: `%s`", l.Manifest.Type)

		name := fmt.Sprintf("[%s](%s) (`/%s`)",
			l.Manifest.DisplayName, l.Manifest.HomepageURL, l.Manifest.AppID)
		txt += md.Markdownf("|%s|%s|%s|%s|%s|%s|\n",
			name, status, version, "", l.Manifest.RequestedLocations, l.Manifest.RequestedPermissions)
	}

	return out(params, txt)
}
