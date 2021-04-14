// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (p *Proxy) InstallApp(sessionID, actingUserID string, cc *apps.Context, trusted bool, secret string) (*apps.App, md.MD, error) {
	session, err := utils.LoadSession(p.mm, sessionID, actingUserID)
	if err != nil {
		return nil, "", err
	}

	err = utils.EnsureSysAdmin(p.mm, actingUserID)
	if err != nil {
		return nil, "", err
	}

	m, err := p.store.Manifest.Get(cc.AppID)
	if err != nil {
		return nil, "", err
	}

	app, err := p.store.App.Get(cc.AppID)
	if err != nil {
		if errors.Cause(err) != utils.ErrNotFound {
			return nil, "", err
		}
		app = &apps.App{}
	}

	app.Manifest = *m
	if app.Disabled {
		app.Disabled = false
	}
	app.GrantedPermissions = m.RequestedPermissions
	app.GrantedLocations = m.RequestedLocations
	if secret != "" {
		app.Secret = secret
	}

	if app.GrantedPermissions.Contains(apps.PermissionRemoteWebhooks) {
		app.WebhookSecret = model.NewId()
	}

	conf := p.conf.GetConfig()
	asAdmin := model.NewAPIv4Client(conf.MattermostSiteURL)
	asAdmin.SetToken(session.Token)

	if app.GrantedPermissions.Contains(apps.PermissionActAsBot) {
		var bot *model.Bot
		var token *model.UserAccessToken
		bot, token, err = p.ensureBot(m, cc.ActingUserID, asAdmin)
		if err != nil {
			return nil, "", err
		}

		app.BotUserID = bot.UserId
		app.BotUsername = bot.Username
		app.BotAccessToken = token.Token
	}

	if app.GrantedPermissions.Contains(apps.PermissionActAsUser) {
		var oAuthApp *model.OAuthApp
		oAuthApp, err = p.ensureOAuthApp(app, trusted, cc.ActingUserID, asAdmin)
		if err != nil {
			return nil, "", err
		}
		app.MattermostOAuth2.ClientID = oAuthApp.Id
		app.MattermostOAuth2.ClientSecret = oAuthApp.ClientSecret
		app.Trusted = trusted
	}

	err = p.store.App.Save(app)
	if err != nil {
		return nil, "", err
	}

	creq := &apps.CallRequest{
		Call:    *apps.DefaultOnInstall.WithOverrides(app.OnInstall),
		Context: cc,
	}
	resp := p.Call(sessionID, actingUserID, creq)
	if resp.Type == apps.CallResponseTypeError {
		return nil, "", errors.Wrap(resp, "install failed")
	}

	p.mm.Frontend.PublishWebSocketEvent(config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{UserId: cc.ActingUserID})
	return app, resp.Markdown, nil
}

func (p *Proxy) ensureOAuthApp(app *apps.App, noUserConsent bool, actingUserID string, asAdmin *model.Client4) (*model.OAuthApp, error) {
	if app.MattermostOAuth2.ClientID != "" {
		oauthApp, response := asAdmin.GetOAuthApp(app.MattermostOAuth2.ClientID)
		if response.StatusCode == http.StatusOK && response.Error == nil {
			_ = p.mm.Post.DM(app.BotUserID, actingUserID, &model.Post{
				Message: fmt.Sprintf("Using existing OAuth2 App `%s`.", oauthApp.Id),
			})

			return oauthApp, nil
		}
	}

	oauth2CallbackURL := p.conf.GetConfig().AppPath(app.AppID) + config.PathMattermostOAuth2Complete

	// For the POC this should work, but for the final product I would opt for a RPC method to register the App
	oauthApp, response := asAdmin.CreateOAuthApp(&model.OAuthApp{
		CreatorId:    actingUserID,
		Name:         app.DisplayName,
		Description:  app.Description,
		CallbackUrls: []string{oauth2CallbackURL},
		Homepage:     app.HomepageURL,
		IsTrusted:    noUserConsent,
	})
	if response.StatusCode != http.StatusCreated {
		if response.Error != nil {
			return nil, errors.Wrap(response.Error, "failed to create OAuth2 App")
		}
		return nil, errors.Errorf("failed to create OAuth2 App: received status code %v", response.StatusCode)
	}

	_ = p.mm.Post.DM(app.BotUserID, actingUserID, &model.Post{
		Message: fmt.Sprintf("Created OAuth2 App (`%s`).", oauthApp.Id),
	})

	return oauthApp, nil
}

func (p *Proxy) ensureBot(manifest *apps.Manifest, actingUserID string, client *model.Client4) (*model.Bot, *model.UserAccessToken, error) {
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

	user, res := client.GetUser(fullBot.UserId, "")
	if user == nil {
		return nil, nil, errors.Wrap(res.Error, "failed to get bot user")
	}

	if !strings.Contains(user.Roles, model.SYSTEM_POST_ALL_ROLE_ID) {
		newRoles := user.Roles + " " + model.SYSTEM_POST_ALL_ROLE_ID
		updated, res := client.UpdateUserRoles(fullBot.UserId, newRoles)
		if !updated {
			return nil, nil, errors.Wrap(res.Error, "failed to update bot user's roles")
		}
	}

	token, response := client.CreateUserAccessToken(fullBot.UserId, "Mattermost App Token")
	if response.StatusCode != http.StatusOK {
		if response.Error != nil {
			return nil, nil, response.Error
		}
		return nil, nil, fmt.Errorf("could not create token, status code = %v", response.StatusCode)
	}

	_ = p.mm.Post.DM(fullBot.UserId, actingUserID, &model.Post{
		Message: fmt.Sprintf("Using bot account @%s (`%s`).",
			fullBot.Username, fullBot.UserId),
	})

	return fullBot, token, nil
}
