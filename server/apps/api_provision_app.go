// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type InProvisionApp struct {
	ManifestURL  string
	Context      CallContext
	SessionToken string
	AppSecret    string
	Force        bool
}

func (s *Service) ProvisionApp(in *InProvisionApp) (*App, md.MD, error) {
	manifest, err := s.Client.GetManifest(in.ManifestURL)
	if err != nil {
		return nil, "", err
	}
	if manifest.AppID == "" {
		return nil, "", errors.New("app ID must not be empty")
	}
	_, err = s.Registry.Get(manifest.AppID)
	if err != utils.ErrNotFound && !in.Force {
		return nil, "", errors.Errorf("app %s already provisioned, use Force to overwrite", manifest.AppID)
	}

	// TODO check if acting user is a sysadmin

	conf := s.Configurator.GetConfig()
	client := model.NewAPIv4Client(conf.MattermostSiteURL)
	client.SetToken(in.SessionToken)

	bot, token, err := s.ensureBot(client, manifest, in.Context.ActingUserID)
	if err != nil {
		return nil, "", err
	}

	oAuthApp, err := s.ensureOAuthApp(client, manifest, in.Context.ActingUserID)
	if err != nil {
		return nil, "", err
	}

	app := &App{
		Manifest:               manifest,
		BotUserID:              bot.UserId,
		BotUsername:            bot.Username,
		BotPersonalAccessToken: token.Token,
		OAuthAppID:             oAuthApp.Id,
		OAuthSecret:            oAuthApp.ClientSecret,
		Secret:                 in.AppSecret,
	}
	err = s.Registry.Store(app)
	if err != nil {
		return nil, "", err
	}

	md := md.Markdownf("Provisioned App %s [%s](%s). Bot user @%s.", app)

	return app, md, nil
}

func (s *Service) ensureBot(client *model.Client4, manifest *Manifest, actingUserID string) (*model.Bot, *model.UserAccessToken, error) {
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

	token, response := client.CreateUserAccessToken(fullBot.UserId, "Mattermost App Token")
	if response.StatusCode != http.StatusOK {
		if response.Error != nil {
			return nil, nil, response.Error
		}
		return nil, nil, fmt.Errorf("could not create token, status code = %v", response.StatusCode)
	}

	_ = s.Mattermost.Post.DM(fullBot.UserId, actingUserID, &model.Post{
		Message: "Mattermost bot account has been provisioned. Welcome!",
	})

	return fullBot, token, nil
}

func (s *Service) ensureOAuthApp(client *model.Client4, manifest *Manifest, actingUserID string) (*model.OAuthApp, error) {
	storedApp, err := s.Registry.Get(manifest.AppID)
	if err != nil && err != utils.ErrNotFound {
		return nil, err
	}

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
