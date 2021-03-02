// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package admin

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (adm *Admin) InstallApp(cc *apps.Context, sessionToken apps.SessionToken, in *apps.InInstallApp) (*apps.App, md.MD, error) {
	// TODO <><> check if acting user is a sysadmin

	app, err := adm.store.App().Get(cc.AppID)
	if err != nil {
		return nil, "", err
	}

	app.GrantedPermissions = in.GrantedPermissions
	app.GrantedLocations = in.GrantedLocations
	if in.AppSecret != "" {
		app.Secret = in.AppSecret
	}

	conf := adm.conf.GetConfig()
	client := model.NewAPIv4Client(conf.MattermostSiteURL)
	client.SetToken(string(sessionToken))

	if in.GrantedPermissions.Contains(apps.PermissionActAsUser) {
		var oAuthApp *model.OAuthApp
		oAuthApp, err = adm.ensureOAuthApp(app.Manifest, in.OAuth2TrustedApp, cc.ActingUserID, string(sessionToken))
		if err != nil {
			return nil, "", err
		}
		app.OAuth2ClientID = oAuthApp.Id
		app.OAuth2ClientSecret = oAuthApp.ClientSecret
		app.OAuth2TrustedApp = in.OAuth2TrustedApp

		// Installed app is automatically enabled, since config is done in the installation process
		if app.Status == "" || app.Status == apps.AppStatusRegistered {
			app.Status = apps.AppStatusInstalled
		}
	}

	err = adm.store.App().Save(app)
	if err != nil {
		return nil, "", err
	}

	install := app.Manifest.OnInstall
	if install == nil {
		install = apps.DefaultInstallCall
	}
	install.Values = map[string]interface{}{
		apps.PropOAuth2ClientSecret: app.OAuth2ClientSecret,
	}
	install.Context = cc

	resp := adm.proxy.Call(sessionToken, install)
	if resp.Type == apps.CallResponseTypeError {
		return nil, "", errors.Wrap(resp, "install failed")
	}

	adm.mm.Frontend.PublishWebSocketEvent(api.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{UserId: cc.ActingUserID})
	return app, resp.Markdown, nil
}

func (adm *Admin) ensureOAuthApp(manifest *apps.Manifest, noUserConsent bool, actingUserID, sessionToken string) (*model.OAuthApp, error) {
	app, err := adm.store.App().Get(manifest.AppID)
	if err != nil && err != utils.ErrNotFound {
		return nil, err
	}

	conf := adm.conf.GetConfig()
	client := model.NewAPIv4Client(conf.MattermostSiteURL)
	client.SetToken(sessionToken)

	if app.OAuth2ClientID != "" {
		oauthApp, response := client.GetOAuthApp(app.OAuth2ClientID)
		if response.StatusCode == http.StatusOK && response.Error == nil {
			_ = adm.mm.Post.DM(app.BotUserID, actingUserID, &model.Post{
				Message: fmt.Sprintf("Using existing OAuth2 App `%s`.", oauthApp.Id),
			})

			return oauthApp, nil
		}
	}

	oauth2CallbackURL := adm.conf.GetConfig().PluginURL + api.AppsPath + "/" + string(manifest.AppID) + api.PathOAuth2Complete

	// For the POC this should work, but for the final product I would opt for a RPC method to register the App
	oauthApp, response := client.CreateOAuthApp(&model.OAuthApp{
		CreatorId:    actingUserID,
		Name:         manifest.DisplayName,
		Description:  manifest.Description,
		CallbackUrls: []string{oauth2CallbackURL},
		Homepage:     manifest.HomepageURL,
		IsTrusted:    noUserConsent,
	})
	if response.StatusCode != http.StatusCreated {
		if response.Error != nil {
			return nil, errors.Wrap(response.Error, "failed to create OAuth2 App")
		}
		return nil, errors.Errorf("failed to create OAuth2 App: received status code %v", response.StatusCode)
	}

	_ = adm.mm.Post.DM(app.BotUserID, actingUserID, &model.Post{
		Message: fmt.Sprintf("Created OAuth2 App (`%s`).", oauthApp.Id),
	})

	return oauthApp, nil
}
