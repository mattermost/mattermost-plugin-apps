// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (a *builtinApp) listCommandBinding() *apps.Binding {
	return commandBinding("list", pList, "[ flags ]", "Display available and installed Apps")
}

// func (a *builtinApp) listForm(c *apps.Call) *apps.CallResponse {
// 	return &apps.CallResponse{
// 		Type: apps.CallResponseTypeForm,
// 		Form: &apps.Form{
// 			Title: "list Apps",
// 			Call: &apps.Call{
// 				URL: PathList,
// 			},
// 		},
// 	}
// }

func (a *builtinApp) list(creq *apps.CallRequest) *apps.CallResponse {
	// marketplaceApps := a.proxy.GetListedApps(filter, includePlugins)
	// installedApps := a.proxy.ListInstalledApps()

	// txt := md.MD("| Name | Status | Version | Account | Locations | Permissions |\n")
	// txt += md.MD("| :-- |:-- | :-- | :-- | :-- | :-- |\n")

	// for _, app := range installedApps {
	// 	mapp := marketplaceApps[app.AppID]
	// 	if mapp == nil {
	// 		continue
	// 	}

	// 	status := "Installed"
	// 	if app.Disabled {
	// 		status += ", Disabled"
	// 	}
	// 	status += fmt.Sprintf(", type: `%s`", app.Type)

	// 	version := string(app.Version)
	// 	if mapp != nil && string(mapp.Manifest.Version) != version {
	// 		version += fmt.Sprintf("(marketplace: %s)", mapp.Manifest.Version)
	// 	}

	// 	account := ""
	// 	if app.BotUserID != "" {
	// 		account += fmt.Sprintf("Bot: `%s`", app.BotUserID)
	// 	}
	// 	if app.OAuth2ClientID != "" {
	// 		if account != "" {
	// 			account += ", "
	// 		}
	// 		account += fmt.Sprintf("OAuth: `%s`", app.OAuth2ClientID)
	// 	}
	// 	name := fmt.Sprintf("[%s](%s) (%s)",
	// 		app.DisplayName, app.HomepageURL, app.AppID)

	// 	txt += md.Markdownf("|%s|%s|%s|%s|%s|%s|\n",
	// 		name, status, version, account, app.GrantedLocations, app.GrantedPermissions)
	// }

	// for _, mapp := range marketplaceApps {
	// 	_, ok := installedApps[mapp.Manifest.AppID]
	// 	if ok {
	// 		continue
	// 	}

	// 	version := string(mapp.Manifest.Version)
	// 	status := fmt.Sprintf("type: `%s`", mapp.Manifest.Type)

	// 	name := fmt.Sprintf("[%s](%s) (%s)",
	// 		mapp.Manifest.DisplayName, mapp.Manifest.HomepageURL, mapp.Manifest.AppID)
	// 	txt += md.Markdownf("|%s|%s|%s|%s|%s|%s|\n",
	// 		name, status, version, "", mapp.Manifest.RequestedLocations, mapp.Manifest.RequestedPermissions)
	// }

	// return apps.NewCallResponse(txt, nil, nil)
	return nil
}
