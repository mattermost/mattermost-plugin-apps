// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
	"github.com/mattermost/mattermost-server/v5/model"
)

type InInstallApp struct {
	ActingMattermostUserID string
	App                    *App
	LogChannelID           string
	LogRootPostID          string
	SessionToken           string
}

type OutInstallApp struct {
	md.MD
	App *App
}

func (s *Service) InstallApp(in *InInstallApp) (*OutInstallApp, error) {
	if in.App.Manifest.AppID == "" {
		return nil, errors.New("app ID must not be empty")
	}
	// TODO check if acting user is a sysadmin

	// TODO remove mock, implement for real

	conf := s.Configurator.GetConfig()
	bot, token, err := createBot(conf.MattermostSiteURL, in.SessionToken, in.App.Manifest)
	if err != nil {
		return nil, err
	}
	oAuthApp, err := createOAuthApp(conf.MattermostSiteURL, in.SessionToken, in.ActingMattermostUserID, in.App.Manifest)
	if err != nil {
		return nil, err
	}

	app := *in.App
	app.GrantedPermissions = app.Manifest.RequestedPermissions
	app.BotID = bot.UserId
	app.BotToken = token.Token
	app.OAuthAppID = oAuthApp.Id
	app.OAuthSecret = oAuthApp.ClientSecret

	err = s.Registry.Store(&app)
	if err != nil {
		return nil, err
	}

	// TODO expand CallData
	callData := &CallData{
		Values: FormValues{},
		Env: map[string]interface{}{
			"log_root_post_id": in.LogRootPostID,
			"log_channel_id":   in.LogChannelID,
		},
		Expanded: &Expanded{
			App: &app,
		},
	}
	callData.Expanded.App.Manifest = nil
	callData.Expanded.App.Secret = ""

	resp, err := s.PostWish(app.Manifest.AppID, in.ActingMattermostUserID, app.Manifest.Install, callData)
	if err != nil {
		return nil, errors.Wrap(err, "Install failed")
	}

	out := &OutInstallApp{
		MD:  resp.Markdown,
		App: &app,
	}
	return out, nil
}

func createBot(mattermostSiteURL, sessionToken string, manifest *Manifest) (*model.Bot, *model.UserAccessToken, error) {
	client := model.NewAPIv4Client(mattermostSiteURL)
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

func createOAuthApp(mattermostSiteURL, sessionToken, userID string, manifest *Manifest) (*model.OAuthApp, error) {
	// For the POC this should work, but for the final product I would opt for a RPC method to register the App
	client := model.NewAPIv4Client(mattermostSiteURL)
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
