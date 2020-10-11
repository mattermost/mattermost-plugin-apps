// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
	"github.com/mattermost/mattermost-server/v5/model"
)

type InInstallApp struct {
	GrantedPermissions     store.Permissions
	AppSecret              string
	NoUserConsentForOAuth2 bool
}

func (s *service) InstallApp(in *InInstallApp, cc *Context, sessionToken SessionToken) (*store.App, md.MD, error) {
	// TODO check if acting user is a sysadmin
	app, err := s.Store.GetApp(cc.AppID)
	if err != nil {
		return nil, "", err
	}

	app.GrantedPermissions = in.GrantedPermissions
	if in.AppSecret != "" {
		app.Secret = in.AppSecret
	}

	conf := s.Configurator.GetConfig()
	client := model.NewAPIv4Client(conf.MattermostSiteURL)
	client.SetToken(string(sessionToken))

	oAuthApp, err := s.ensureOAuthApp(app.Manifest, in.NoUserConsentForOAuth2, cc.ActingUserID, string(sessionToken))
	if err != nil {
		return nil, "", err
	}
	app.OAuth2ClientID = oAuthApp.Id
	app.OAuth2ClientSecret = oAuthApp.ClientSecret

	err = s.Store.StoreApp(app)
	if err != nil {
		return nil, "", err
	}

	expandedContext, err := s.newExpander(cc).Expand(
		&store.Expand{
			App:    store.ExpandAll,
			Config: true,
		},
	)
	if err != nil {
		return nil, "", err
	}

	fmt.Printf("<><> InstallApp: %#v\n", expandedContext.App)

	resp, err := s.PostWish(
		&Call{
			Wish: app.Manifest.Install,
			Request: &CallRequest{
				Values: FormValues{
					Data: map[string]interface{}{
						"bot_access_token":     app.BotAccessToken,
						"oauth2_client_secret": app.OAuth2ClientSecret},
				},
				Context: expandedContext,
			},
		})
	if err != nil {
		return nil, "", errors.Wrap(err, "Install failed")
	}

	return app, resp.Markdown, nil
}

func (s *service) ensureOAuthApp(manifest *store.Manifest, noUserConsent bool, actingUserID, sessionToken string) (*model.OAuthApp, error) {
	app, err := s.Store.GetApp(manifest.AppID)
	if err != nil && err != utils.ErrNotFound {
		return nil, err
	}

	conf := s.Configurator.GetConfig()
	client := model.NewAPIv4Client(conf.MattermostSiteURL)
	client.SetToken(sessionToken)

	if app.OAuth2ClientID != "" {
		oauthApp, response := client.GetOAuthApp(app.OAuth2ClientID)
		if response.StatusCode == http.StatusOK && response.Error == nil {
			_ = s.Mattermost.Post.DM(app.BotUserID, actingUserID, &model.Post{
				Message: fmt.Sprintf("Using existing OAuth2 App `%s`.", oauthApp.Id),
			})

			return oauthApp, nil
		}
	}

	// For the POC this should work, but for the final product I would opt for a RPC method to register the App
	oauthApp, response := client.CreateOAuthApp(&model.OAuthApp{
		CreatorId:    actingUserID,
		Name:         manifest.DisplayName,
		Description:  manifest.Description,
		CallbackUrls: []string{manifest.OAuth2CallbackURL},
		Homepage:     manifest.HomepageURL,
		IsTrusted:    noUserConsent,
	})
	if response.StatusCode != http.StatusCreated {
		if response.Error != nil {
			return nil, errors.Wrap(response.Error, "failed to create OAuth2 App")
		}
		return nil, errors.Errorf("failed to create OAuth2 App: received status code %v", response.StatusCode)
	}

	_ = s.Mattermost.Post.DM(app.BotUserID, actingUserID, &model.Post{
		Message: fmt.Sprintf("Created OAuth2 App (`%s`).", oauthApp.Id),
	})

	return oauthApp, nil
}
