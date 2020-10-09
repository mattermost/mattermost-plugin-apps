// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
	"github.com/mattermost/mattermost-server/v5/model"
)

type InInstallApp struct {
	GrantedPermissions     Permissions
	AppSecret              string
	NoUserConsentForOAuth2 bool
}

func (s *Service) InstallApp(in *InInstallApp, cc *CallContext, sessionToken SessionToken) (*App, md.MD, error) {
	// TODO check if acting user is a sysadmin
	app, err := s.Registry.Get(cc.AppID)
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

	oAuthApp, err := s.ensureOAuthApp(app.Manifest, cc.ActingUserID, string(sessionToken))
	if err != nil {
		return nil, "", err
	}
	app.OAuthAppID = oAuthApp.Id
	app.OAuthSecret = oAuthApp.ClientSecret

	err = s.Registry.Store(app)
	if err != nil {
		return nil, "", err
	}

	cloneContext := *cc
	cloneContext.AppID = app.Manifest.AppID

	cloneApp := *app
	cloneApp.Manifest = nil
	cloneApp.Secret = ""

	resp, err := s.PostWish(
		Call{
			Wish: app.Manifest.Install,
			Data: &CallData{
				Values:  FormValues{},
				Context: cloneContext,
				Expanded: &Expanded{
					App: &cloneApp,
				},
			},
		})
	if err != nil {
		return nil, "", errors.Wrap(err, "Install failed")
	}

	return app, resp.Markdown, nil
}

func (s *Service) ensureOAuthApp(manifest *Manifest, actingUserID, sessionToken string) (*model.OAuthApp, error) {
	storedApp, err := s.Registry.Get(manifest.AppID)
	if err != nil && err != utils.ErrNotFound {
		return nil, err
	}

	conf := s.Configurator.GetConfig()
	client := model.NewAPIv4Client(conf.MattermostSiteURL)
	client.SetToken(sessionToken)

	if storedApp != nil && storedApp.OAuthAppID != "" {
		oauthApp, response := client.GetOAuthApp(storedApp.OAuthAppID)
		if response.StatusCode == http.StatusOK && response.Error == nil {
			return oauthApp, nil
		}
	}

	// For the POC this should work, but for the final product I would opt for a RPC method to register the App
	oauthApp, response := client.CreateOAuthApp(&model.OAuthApp{
		CreatorId:    actingUserID,
		Name:         manifest.DisplayName,
		Description:  manifest.Description,
		CallbackUrls: []string{manifest.CallbackURL},
		Homepage:     manifest.Homepage,
		IsTrusted:    true,
	})
	if response.StatusCode != http.StatusCreated {
		if response.Error != nil {
			return nil, errors.Wrap(response.Error, "failed to create OAuth2 App")
		}
		return nil, errors.Errorf("failed to create OAuth2 App: received status code %v", response.StatusCode)
	}

	return oauthApp, nil
}
