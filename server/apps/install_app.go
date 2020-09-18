// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
	"github.com/mattermost/mattermost-server/v5/model"
)

type InInstallApp struct {
	ActingMattermostUserID string
	NoUserConsentForOAuth2 bool
	Manifest               *Manifest
	Secret                 string
	SessionToken           string
}

type OutInstallApp struct {
	md.MD
	App *App
}

type InstallRequest struct {
	OAuthAPIKey string
	OAuthSecret string
}

func (r *registry) InstallApp(in *InInstallApp) (*OutInstallApp, error) {
	if in.Manifest.AppID == "" {
		return nil, errors.New("app ID must not be empty")
	}
	// TODO check if acting user is a sysadmin

	// TODO remove mock, implement for real

	bot, token, err := r.createBot(in.Manifest, in.SessionToken)
	if err != nil {
		return nil, err
	}
	oAuthApp, err := r.createOAuthApp(in.ActingMattermostUserID, token.Token, in.Manifest)
	if err != nil {
		return nil, err
	}

	app := &App{
		Manifest:               in.Manifest,
		GrantedPermissions:     in.Manifest.RequestedPermissions,
		NoUserConsentForOAuth2: in.NoUserConsentForOAuth2,
		Secret:                 in.Secret,
		BotID:                  bot.UserId,
		BotToken:               token.Token,
		OAuthAppID:             oAuthApp.Id,
		OAuthSecret:            oAuthApp.ClientSecret,
	}
	r.apps[in.Manifest.AppID] = app

	extraInfo := fmt.Sprintf(`Bot token: %s
	OAuth Client ID: %s
	OAuth Client Secret: %s`, token.Token, oAuthApp.Id, oAuthApp.ClientSecret)

	out := &OutInstallApp{
		MD:  md.Markdownf("Installed %s (%s)\n%s", in.Manifest.DisplayName, in.Manifest.AppID, extraInfo),
		App: app,
	}
	return out, nil
}

func (r *registry) createBot(manifest *Manifest, sessionToken string) (*model.Bot, *model.UserAccessToken, error) {
	client := model.NewAPIv4Client(r.configurator.GetConfig().MattermostSiteURL)
	fmt.Printf("session token: %s\n", sessionToken)
	client.SetToken(sessionToken)
	bot := &model.Bot{
		Username:    string(manifest.AppID),
		DisplayName: manifest.DisplayName,
		Description: fmt.Sprintf("Bot account for `%s` App.", manifest.DisplayName),
	}

	fullBot, response := client.CreateBot(bot)

	if response.StatusCode != http.StatusCreated {
		if response.Error != nil {
			return nil, nil, response.Error
		}
		return nil, nil, errors.New("could not create bot")
	}

	token, response := client.CreateUserAccessToken(fullBot.UserId, "Default Token")
	if response.StatusCode != http.StatusCreated {
		if response.Error != nil {
			return nil, nil, response.Error
		}
		return nil, nil, errors.New("could not create token")
	}

	return fullBot, token, nil
}

func (r *registry) createOAuthApp(userID string, sessionToken string, manifest *Manifest) (*model.OAuthApp, error) {
	// For the POC this should work, but for the final product I would opt for a RPC method to register the App
	client := model.NewAPIv4Client(r.configurator.GetConfig().MattermostSiteURL)
	app := model.OAuthApp{
		CreatorId:    userID,
		Name:         manifest.DisplayName,
		Description:  manifest.Description,
		CallbackUrls: []string{},
		IsTrusted:    true,
	}

	fullApp, response := client.CreateOAuthApp(&app)

	if response.StatusCode != http.StatusCreated {
		if response.Error != nil {
			return nil, response.Error
		}
		return nil, errors.New("could not create the app")
	}

	return fullApp, nil
}
