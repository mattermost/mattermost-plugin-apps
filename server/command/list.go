// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package command

import (
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (s *service) executeList(params *params) (*model.CommandResponse, error) {
	marketplaceApps := s.admin.ListMarketplaceApps("")
	installedApps := s.admin.ListInstalledApps()

	txt := md.MD("| Name | Status | Version | Account | Locations | Permissions |\n")
	txt += md.MD("| :-- |:-- | :-- | :-- | :-- | :-- |\n")

	for _, app := range installedApps {
		mapp := marketplaceApps[app.AppID]
		if mapp == nil {
			continue
		}

		status := "**Installed**"
		if app.Disabled {
			status += ", **Disabled**"
		}
		status += fmt.Sprintf(", type: `%s`", app.Type)

		version := string(app.Version)
		if string(mapp.Manifest.Version) != version {
			version += fmt.Sprintf(", %s in marketplace", mapp.Manifest.Version)
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

	for _, mapp := range marketplaceApps {
		_, ok := installedApps[mapp.Manifest.AppID]
		if ok {
			continue
		}

		version := string(mapp.Manifest.Version)
		status := fmt.Sprintf("type: `%s`", mapp.Manifest.Type)

		name := fmt.Sprintf("[%s](%s) (`/%s`)",
			mapp.Manifest.DisplayName, mapp.Manifest.HomepageURL, mapp.Manifest.AppID)
		txt += md.Markdownf("|%s|%s|%s|%s|%s|%s|\n",
			name, status, version, "", mapp.Manifest.RequestedLocations, mapp.Manifest.RequestedPermissions)
	}

	return out(params, txt)
}
