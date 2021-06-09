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
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/md"
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

	conf := p.conf.GetConfig()
	err = isAppTypeSupported(conf, m)
	if err != nil {
		return nil, "", err
	}

	app, err := p.store.App.Get(cc.AppID)
	if err != nil {
		if !errors.Is(err, utils.ErrNotFound) {
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

	asAdmin := model.NewAPIv4Client(conf.MattermostSiteURL)
	asAdmin.SetToken(session.Token)

	err = p.ensureBot(app, cc.ActingUserID, asAdmin)
	if err != nil {
		return nil, "", err
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

	var message md.MD
	if app.OnInstall != nil {
		creq := &apps.CallRequest{
			Call:    *apps.DefaultOnInstall.WithOverrides(app.OnInstall),
			Context: cc,
		}
		resp := p.Call(sessionID, actingUserID, creq)
		// TODO fail on all errors except 404
		if resp.Type == apps.CallResponseTypeError {
			p.mm.Log.Warn("OnInstall failed, installing app anyway", "err", resp.Error(), "app_id", app.AppID)
		}

		message = resp.Markdown
	}

	if message == "" {
		message = md.MD(fmt.Sprintf("Successfully installed %s", app.AppID))
	}

	p.dispatchRefreshBindingsEvent(cc.ActingUserID)
	return app, message, nil
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

	oauth2CallbackURL := p.conf.GetConfig().AppURL(app.AppID) + config.PathMattermostOAuth2Complete

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

func (p *Proxy) ensureBot(app *apps.App, actingUserID string, client *model.Client4) error {
	bot := &model.Bot{
		Username:    strings.ToLower(string(app.AppID)),
		DisplayName: app.DisplayName,
		Description: fmt.Sprintf("Bot account for `%s` App.", app.DisplayName),
	}

	var fullBot *model.Bot
	var response *model.Response
	user, _ := client.GetUserByUsername(bot.Username, "")
	if user == nil {
		fullBot, response = client.CreateBot(bot)

		if response.StatusCode != http.StatusCreated {
			if response.Error != nil {
				return response.Error
			}
			return errors.New("could not create bot")
		}
	} else {
		if !user.IsBot {
			return errors.New("a user already owns the bot username")
		}

		fullBot = model.BotFromUser(user)
		if user.DeleteAt != 0 {
			fullBot, response = client.EnableBot(fullBot.UserId)
			if response.StatusCode != http.StatusOK {
				if response.Error != nil {
					return response.Error
				}
				return errors.New("could not enable bot")
			}
		}
	}
	app.BotUserID = fullBot.UserId
	app.BotUsername = fullBot.Username

	err := p.updateBotIcon(app)
	if err != nil {
		return errors.Wrap(err, "failed set bot icon")
	}

	if app.RequestedPermissions.Contains(apps.PermissionActAsBot) {
		var token *model.UserAccessToken
		if app.BotAccessTokenID != "" {
			token, response = client.GetUserAccessToken(app.BotAccessTokenID)
			if response.Error != nil {
				return errors.Wrap(response.Error, "failed to create bot user's access token")
			}
			if response.StatusCode != http.StatusOK {
				return fmt.Errorf("failed to create bot user's access token, status code = %v", response.StatusCode)
			}
		} else {
			token, response = client.CreateUserAccessToken(fullBot.UserId, "Mattermost App Token")
			if response.Error != nil {
				return errors.Wrap(response.Error, "failed to create bot user's access token")
			}
			if response.StatusCode != http.StatusOK {
				return fmt.Errorf("failed to create bot user's access token, status code = %v", response.StatusCode)
			}
		}
		app.BotAccessToken = token.Token
		app.BotAccessTokenID = token.Id
	}

	_ = p.mm.Post.DM(fullBot.UserId, actingUserID, &model.Post{
		Message: fmt.Sprintf("Using bot account @%s (`%s`).",
			fullBot.Username, fullBot.UserId),
	})

	return nil
}

func (p *Proxy) updateBotIcon(app *apps.App) error {
	iconPath := app.Manifest.Icon

	// If app doesn't have an icon, do nothing
	if iconPath == "" {
		return nil
	}

	asset, _, err := p.GetAsset(app.AppID, iconPath)
	if err != nil {
		return errors.Wrap(err, "failed to get app icon")
	}
	defer asset.Close()

	err = p.mm.User.SetProfileImage(app.BotUserID, asset)
	if err != nil {
		return errors.Wrap(err, "update profile icon")
	}

	return nil
}
