// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/md"
)

// InstallApp installs an App.
//  - client is a user-scoped(??) client to Mattermost??
//  - sessionID is needed to pass down to the app in liue of a proper token
//  - cc is the Context that will be passed down to the App's OnInstall callback.
func (p *Proxy) InstallApp(in Incoming, appID apps.AppID, deployType apps.DeployType, trusted bool, secret string) (*apps.App, md.MD, error) {
	m, err := p.store.Manifest.Get(appID)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to find manifest to install app")
	}
	if !m.SupportsDeploy(deployType) {
		return nil, "", errors.Errorf("app does not support %s deployment", deployType)
	}
	err = CanDeploy(p, deployType)
	if err != nil {
		return nil, "", err
	}

	app, err := p.store.App.Get(appID)
	if err != nil {
		if !errors.Is(err, utils.ErrNotFound) {
			return nil, "", errors.Wrap(err, "failed looking for existing app")
		}
		app = &apps.App{}
	}

	app.DeployType = deployType
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

	client := p.newSudoClient(in)
	err = p.ensureBot(client, app)
	if err != nil {
		return nil, "", err
	}

	if app.GrantedPermissions.Contains(apps.PermissionActAsUser) {
		var oAuthApp *model.OAuthApp
		oAuthApp, err = p.ensureOAuthApp(client, app, trusted, in.ActingUserID)
		if err != nil {
			return nil, "", err
		}
		app.MattermostOAuth2.ClientID = oAuthApp.Id
		app.MattermostOAuth2.ClientSecret = oAuthApp.ClientSecret
		app.Trusted = trusted
	}

	err = p.store.App.Save(*app)
	if err != nil {
		return nil, "", err
	}

	var message md.MD
	if app.OnInstall != nil {
		resp := p.simpleCall(in, app, *app.OnInstall)
		// TODO fail on all errors except 404
		if resp.Type == apps.CallResponseTypeError {
			p.log.WithError(err).Warnw("OnInstall failed, installing app anyway", "app_id", app.AppID)
		} else {
			message = resp.Markdown
		}
	}

	if message == "" {
		message = md.MD(fmt.Sprintf("Installed %s", app.DisplayName))
	}

	p.log.Infow("Installed an app",
		"app_id", app.AppID)

	p.dispatchRefreshBindingsEvent(in.ActingUserID)

	return app, message, nil
}

func (p *Proxy) ensureOAuthApp(client mmclient.Client, app *apps.App, noUserConsent bool, actingUserID string) (*model.OAuthApp, error) {
	if app.MattermostOAuth2.ClientID != "" {
		oauthApp, err := client.GetOAuthApp(app.MattermostOAuth2.ClientID)
		if err == nil {
			p.log.Debugw("App install flow: Using existing OAuth2 App",
				"id", oauthApp.Id)

			return oauthApp, nil
		}
	}

	oauth2CallbackURL := p.conf.GetConfig().AppURL(app.AppID) + config.PathMattermostOAuth2Complete

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

	p.log.Debugw("App install flow: Created OAuth2 App",
		"id", oauthApp.Id)

	return oauthApp, nil
}

func (p *Proxy) ensureBot(client mmclient.Client, app *apps.App) error {
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

		p.log.Debugw("App install flow: Created Bot Account ",
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
		bot.UserId = user.Id
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

	asset, _, err := p.getStatic(app, iconPath)
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
