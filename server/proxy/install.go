// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (p *Proxy) InstallApp(client mmclient.Client, sessionID string, cc *apps.Context, trusted bool, secret string) (*apps.App, string, error) {
	conf, _, log := p.conf.Basic()
	log = log.With("app_id", cc.AppID)
	m, err := p.store.Manifest.Get(cc.AppID)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to find manifest to install app")
	}

	err = isAppTypeSupported(conf, m.AppType)
	if err != nil {
		return nil, "", errors.Wrap(err, "app type is not supported")
	}

	app, err := p.store.App.Get(cc.AppID)
	if err != nil {
		if !errors.Is(err, utils.ErrNotFound) {
			return nil, "", errors.Wrap(err, "failed to find existing app")
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

	err = p.ensureBot(client, app)
	if err != nil {
		return nil, "", err
	}

	if app.GrantedPermissions.Contains(apps.PermissionActAsUser) {
		var oAuthApp *model.OAuthApp
		oAuthApp, err = p.ensureOAuthApp(client, app, trusted, cc.ActingUserID)
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

	var message string
	if app.OnInstall != nil {
		creq := &apps.CallRequest{
			Call:    *app.OnInstall,
			Context: cc,
		}
		resp := p.Call(sessionID, cc.ActingUserID, creq)
		// TODO fail on all errors except 404
		if resp.Type == apps.CallResponseTypeError {
			log.WithError(err).Warnf("OnInstall failed, installing app anyway.")
		} else {
			message = resp.Markdown
		}
	}

	if message == "" {
		message = fmt.Sprintf("Installed %s", app.DisplayName)
	}

	log.Infof("Installed app.")

	p.conf.Telemetry().TrackInstall(string(app.AppID), string(app.AppType))

	p.dispatchRefreshBindingsEvent(cc.ActingUserID)

	return app, message, nil
}

func (p *Proxy) ensureOAuthApp(client mmclient.Client, app *apps.App, noUserConsent bool, actingUserID string) (*model.OAuthApp, error) {
	conf, _, log := p.conf.Basic()

	if app.MattermostOAuth2.ClientID != "" {
		oauthApp, err := client.GetOAuthApp(app.MattermostOAuth2.ClientID)
		if err == nil {
			log.Debugw("App install flow: Using existing OAuth2 App",
				"id", oauthApp.Id)
			return oauthApp, nil
		}
	}

	oauth2CallbackURL := conf.AppURL(app.AppID) + config.PathMattermostOAuth2Complete

	oauthApp := &model.OAuthApp{
		CreatorId:    actingUserID,
		Name:         app.DisplayName,
		Description:  app.Description,
		CallbackUrls: []string{oauth2CallbackURL},
		Homepage:     app.HomepageURL,
		IsTrusted:    noUserConsent,
	}
	err := client.CreateOAuthApp(oauthApp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create OAuth2 App")
	}

	log.Debugw("App install flow: Created OAuth2 App", "id", oauthApp.Id)

	return oauthApp, nil
}

func (p *Proxy) ensureBot(client mmclient.Client, app *apps.App) error {
	log := p.conf.Logger()
	bot := &model.Bot{
		Username:    strings.ToLower(string(app.AppID)),
		DisplayName: app.DisplayName,
		Description: fmt.Sprintf("Bot account for `%s` App.", app.DisplayName),
	}

	user, _ := client.GetUserByUsername(bot.Username)
	if user == nil {
		err := client.CreateBot(bot)
		if err != nil {
			return err
		}

		log.Debugw("App install flow: Created Bot Account ",
			"username", bot.Username)
	} else {
		if !user.IsBot {
			return errors.New("a user already owns the bot username")
		}

		// Check if disabled
		if user.DeleteAt != 0 {
			var err error
			bot, err = client.EnableBot(user.Id)
			if err != nil {
				return err
			}
		}

		_, err := client.GetBot(user.Id)
		if err != nil {
			err = client.CreateBot(bot)
			if err != nil {
				return err
			}
		} else {
			bot.UserId = user.Id
			bot.Username = user.Username
		}
	}
	app.BotUserID = bot.UserId
	app.BotUsername = bot.Username

	err := p.updateBotIcon(app)
	if err != nil {
		return errors.Wrap(err, "failed set bot icon")
	}

	// Create an access token on a fresh app install
	if app.RequestedPermissions.Contains(apps.PermissionActAsBot) &&
		app.BotAccessTokenID == "" {
		token, err := client.CreateUserAccessToken(bot.UserId, "Mattermost App Token")
		if err != nil {
			return errors.Wrap(err, "failed to create bot user's access token")
		}

		app.BotAccessToken = token.Token
		app.BotAccessTokenID = token.Id
	}

	return nil
}

func (p *Proxy) updateBotIcon(app *apps.App) error {
	iconPath := app.Manifest.Icon

	// If app doesn't have an icon, do nothing
	if iconPath == "" {
		return nil
	}

	asset, _, err := p.getStatic(&app.Manifest, iconPath)
	if err != nil {
		return errors.Wrap(err, "failed to get app icon")
	}
	defer asset.Close()

	err = p.conf.MattermostAPI().User.SetProfileImage(app.BotUserID, asset)
	if err != nil {
		return errors.Wrap(err, "update profile icon")
	}

	return nil
}
