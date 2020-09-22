// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/appmodel"
	"github.com/mattermost/mattermost-plugin-apps/server/client"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
	"github.com/mattermost/mattermost-server/v5/model"
)

type InInstallApp struct {
	ActingMattermostUserID string
	NoUserConsentForOAuth2 bool
	Manifest               *appmodel.Manifest
	Secret                 string
	SessionToken           string
}

type OutInstallApp struct {
	md.MD
	App *appmodel.App
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
	oAuthApp, err := r.createOAuthApp(in.ActingMattermostUserID, in.SessionToken, in.Manifest)
	if err != nil {
		return nil, err
	}

	app := &appmodel.App{
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

	c, err := client.New(in.ActingMattermostUserID, app)
	if err != nil {
		return nil, err
	}

	c.InstallComplete()

	out := &OutInstallApp{
		MD:  md.Markdownf("Installed %s (%s)", in.Manifest.DisplayName, in.Manifest.AppID),
		App: app,
	}

	return out, nil
}

func (r *registry) createBot(manifest *appmodel.Manifest, sessionToken string) (*model.Bot, *model.UserAccessToken, error) {
	client := model.NewAPIv4Client(r.configurator.GetConfig().MattermostSiteURL)
	client.SetToken(sessionToken)
	bot := &model.Bot{
		Username:    string(manifest.AppID),
		DisplayName: manifest.DisplayName,
		Description: fmt.Sprintf("Bot account for `%s` App.", manifest.DisplayName),
	}

	var fullBot *model.Bot
	user, _ := client.GetUserByUsername(bot.Username, "")
	if user == nil {
		var response *model.Response
		fullBot, response = client.CreateBot(bot)

		if response.StatusCode != http.StatusCreated {
			if response.Error != nil {
				return nil, nil, response.Error
			}
			return nil, nil, errors.New("could not create bot")
		}
	} else {
		if !user.IsBot {
			return nil, nil, errors.New("a user already owns the bot username")
		}

		fullBot = model.BotFromUser(user)
		if fullBot.DeleteAt != 0 {
			var response *model.Response
			fullBot, response = client.EnableBot(fullBot.UserId)
			if response.StatusCode != http.StatusOK {
				if response.Error != nil {
					return nil, nil, response.Error
				}
				return nil, nil, errors.New("could not enable bot")
			}
		}
	}

	token, response := client.CreateUserAccessToken(fullBot.UserId, "Default Token")
	if response.StatusCode != http.StatusOK {
		if response.Error != nil {
			return nil, nil, response.Error
		}
		return nil, nil, fmt.Errorf("could not create token, status code = %v", response.StatusCode)
	}

	return fullBot, token, nil
}

func (r *registry) createOAuthApp(userID string, sessionToken string, manifest *appmodel.Manifest) (*model.OAuthApp, error) {
	// For the POC this should work, but for the final product I would opt for a RPC method to register the App
	client := model.NewAPIv4Client(r.configurator.GetConfig().MattermostSiteURL)
	app := model.OAuthApp{
		CreatorId:    userID,
		Name:         manifest.DisplayName,
		Description:  manifest.Description,
		CallbackUrls: []string{manifest.CallbackURL},
		Homepage:     manifest.Homepage,
		IsTrusted:    true,
	}

	client.SetToken(sessionToken)

	fullApp, response := client.CreateOAuthApp(&app)

	if response.StatusCode != http.StatusCreated {
		if response.Error != nil {
			return nil, fmt.Errorf("error creating the app, %v", response.Error)
		}
		return nil, fmt.Errorf("could not create the app, status code = %v", response.StatusCode)
	}

	return fullApp, nil
}
