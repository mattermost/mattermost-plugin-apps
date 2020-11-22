// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package admin

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
	"github.com/mattermost/mattermost-server/v5/model"
)

func (a *Admin) InstallApp(cc *api.Context, sessionToken api.SessionToken, in *api.InInstallApp) (*api.App, md.MD, error) {
	// TODO <><> check if acting user is a sysadmin
	app, err := a.store.LoadApp(cc.AppID)
	if err != nil {
		return nil, "", err
	}

	app.GrantedPermissions = in.GrantedPermissions
	app.GrantedLocations = in.GrantedLocations
	if in.AppSecret != "" {
		app.Secret = in.AppSecret
	}

	conf := a.conf.GetConfig()
	client := model.NewAPIv4Client(conf.MattermostSiteURL)
	client.SetToken(string(sessionToken))

	oAuthApp, err := a.ensureOAuthApp(app.Manifest, in.OAuth2TrustedApp, cc.ActingUserID, string(sessionToken))
	if err != nil {
		return nil, "", err
	}
	app.OAuth2ClientID = oAuthApp.Id
	app.OAuth2ClientSecret = oAuthApp.ClientSecret
	app.OAuth2TrustedApp = in.OAuth2TrustedApp

	err = a.store.StoreApp(app)
	if err != nil {
		return nil, "", err
	}

	install := app.Manifest.Install
	if install == nil {
		install = api.DefaultInstall
	}
	install.Values = map[string]string{
		api.PropBotAccessToken:     app.BotAccessToken,
		api.PropOAuth2ClientSecret: app.OAuth2ClientSecret,
	}
	install.Context = cc

	resp, err := a.proxy.Call(install)
	if err != nil {
		return nil, "", errors.Wrap(err, "Install failed")
	}
	if resp.Error != "" {
		return nil, "", errors.New("Install failed: " + resp.Error)
	}

	return app, resp.Markdown, nil
}

func (a *Admin) ensureOAuthApp(manifest *api.Manifest, noUserConsent bool, actingUserID, sessionToken string) (*model.OAuthApp, error) {
	app, err := a.store.LoadApp(manifest.AppID)
	if err != nil && err != utils.ErrNotFound {
		return nil, err
	}

	conf := a.conf.GetConfig()
	client := model.NewAPIv4Client(conf.MattermostSiteURL)
	client.SetToken(sessionToken)

	if app.OAuth2ClientID != "" {
		oauthApp, response := client.GetOAuthApp(app.OAuth2ClientID)
		if response.StatusCode == http.StatusOK && response.Error == nil {
			_ = a.mm.Post.DM(app.BotUserID, actingUserID, &model.Post{
				Message: fmt.Sprintf("<><> Using existing OAuth2 App `%s`.", oauthApp.Id),
			})

			return oauthApp, nil
		}
	}

	// For the POC this should work, but for the final product I would opt for a RPC method to register the App
	oauthApp, response := client.CreateOAuthApp(&model.OAuthApp{
		CreatorId:    actingUserID,
		Name:         manifest.DisplayName,
		Description:  manifest.Description,
		CallbackUrls: []string{manifest.RemoteOAuth2CallbackURL},
		Homepage:     manifest.HomepageURL,
		IsTrusted:    noUserConsent,
	})
	if response.StatusCode != http.StatusCreated {
		if response.Error != nil {
			return nil, errors.Wrap(response.Error, "failed to create OAuth2 App")
		}
		return nil, errors.Errorf("failed to create OAuth2 App: received status code %v", response.StatusCode)
	}

	_ = a.mm.Post.DM(app.BotUserID, actingUserID, &model.Post{
		Message: fmt.Sprintf("Created OAuth2 App (`%s`).", oauthApp.Id),
	})

	return oauthApp, nil
}
