// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type InInstallApp struct {
	Context      CallContext
	App          App
	SessionToken string
}

type OutInstallApp struct {
	md.MD
	App *App
}

func (s *Service) InstallApp(in InInstallApp) (*OutInstallApp, error) {
	if in.App.Manifest.AppID == "" {
		return nil, errors.New("app ID must not be empty")
	}
	// TODO check if acting user is a sysadmin

	conf := s.Configurator.GetConfig()
	client := model.NewAPIv4Client(conf.MattermostSiteURL)
	client.SetToken(in.SessionToken)
	bot, token, err := createBot(client, &in)
	if err != nil {
		return nil, err
	}
	oAuthApp, err := createOAuthApp(client, &in)
	if err != nil {
		return nil, err
	}

	app := in.App
	app.GrantedPermissions = app.Manifest.RequestedPermissions
	app.BotUserID = bot.UserId
	app.BotPersonalAccessToken = token.Token
	app.OAuthAppID = oAuthApp.Id
	app.OAuthSecret = oAuthApp.ClientSecret

	err = s.Registry.Store(&app)
	if err != nil {
		return nil, err
	}

	in.Context.AppID = app.Manifest.AppID
	expApp := app
	expApp.Manifest = nil
	expApp.Secret = ""

	resp, err := s.PostWish(
		Call{
			Wish: app.Manifest.Install,
			Data: &CallData{
				Values:  FormValues{},
				Context: in.Context,
				Expanded: &Expanded{
					App: &expApp,
				},
			},
		})
	if err != nil {
		return nil, errors.Wrap(err, "Install failed")
	}

	out := &OutInstallApp{
		MD:  resp.Markdown,
		App: &app,
	}
	return out, nil
}

func createBot(client *model.Client4, in *InInstallApp) (*model.Bot, *model.UserAccessToken, error) {
	bot := &model.Bot{
		Username:    string(in.App.Manifest.AppID),
		DisplayName: in.App.Manifest.DisplayName,
		Description: fmt.Sprintf("Bot account for `%s` App.", in.App.Manifest.DisplayName),
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

func createOAuthApp(client *model.Client4, in *InInstallApp) (*model.OAuthApp, error) {
	// For the POC this should work, but for the final product I would opt for a RPC method to register the App
	m := in.App.Manifest
	app := model.OAuthApp{
		CreatorId:    in.Context.ActingUserID,
		Name:         m.DisplayName,
		Description:  m.Description,
		CallbackUrls: []string{m.CallbackURL},
		Homepage:     m.Homepage,
		IsTrusted:    true,
	}

	fullApp, response := client.CreateOAuthApp(&app)
	if response.StatusCode != http.StatusCreated {
		if response.Error != nil {
			return nil, fmt.Errorf("error creating the app, %v", response.Error)
		}
		return nil, fmt.Errorf("could not create the app, status code = %v", response.StatusCode)
	}

	return fullApp, nil
}
