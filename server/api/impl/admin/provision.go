// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package admin

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (adm *Admin) ProvisionApp(cc *api.Context, sessionToken api.SessionToken, in *api.InProvisionApp) (*api.App, md.MD, error) {
	manifest := in.Manifest
	if manifest.AppID == "" {
		return nil, "", errors.New("app ID must not be empty")
	}
	_, err := adm.store.LoadApp(manifest.AppID)
	if err != utils.ErrNotFound && !in.Force {
		return nil, "", errors.Errorf("app %s already provisioned, use Force to overwrite", manifest.AppID)
	}

	// TODO check if acting user is a sysadmin

	bot, token, err := adm.ensureBot(manifest, cc.ActingUserID, string(sessionToken))
	if err != nil {
		return nil, "", err
	}

	app := &api.App{
		Manifest:       manifest,
		BotUserID:      bot.UserId,
		BotUsername:    bot.Username,
		BotAccessToken: token.Token,
		Secret:         in.AppSecret,
	}
	err = adm.store.StoreApp(app)
	if err != nil {
		return nil, "", err
	}

	md := md.Markdownf("Provisioned App %s [%s](%s). Bot user @%s.",
		app.Manifest.AppID, app.Manifest.DisplayName, app.Manifest.HomepageURL, app.BotUsername)

	return app, md, nil
}

func (adm *Admin) ensureBot(manifest *api.Manifest, actingUserID, sessionToken string) (*model.Bot, *model.UserAccessToken, error) {
	conf := adm.conf.GetConfig()
	client := model.NewAPIv4Client(conf.MattermostSiteURL)
	client.SetToken(sessionToken)

	bot := &model.Bot{
		Username:    strings.ToLower(string(manifest.AppID)),
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

	_ = adm.mm.Post.DM(fullBot.UserId, actingUserID, &model.Post{
		Message: fmt.Sprintf("Provisioned bot account @%s (`%s`).",
			fullBot.Username, fullBot.UserId),
	})

	return fullBot, token, nil
}
