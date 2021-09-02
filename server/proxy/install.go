// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// InstallApp installs an App.
//  - cc is the Context that will be passed down to the App's OnInstall callback.
func (p *Proxy) InstallApp(in Incoming, cc apps.Context, appID apps.AppID, trusted bool, secret string) (*apps.App, string, error) {
	conf, mm, log := p.conf.Basic()
	log = log.With("app_id", appID)
	m, err := p.store.Manifest.Get(appID)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to find manifest to install app")
	}
	err = isAppTypeSupported(conf, m.AppType)
	if err != nil {
		return nil, "", errors.Wrap(err, "app type is not supported")
	}

	app, err := p.store.App.Get(appID)
	if err != nil {
		if !errors.Is(err, utils.ErrNotFound) {
			return nil, "", errors.Wrap(err, "failed looking for existing app")
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

	icon, err := p.getAppIcon(app)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed get bot icon")
	}
	if icon != nil {
		defer icon.Close()
	}

	asAdmin := mmclient.NewRPCClient(mm)
	err = p.ensureBot(asAdmin, log, app, icon)
	if err != nil {
		return nil, "", err
	}

	if app.GrantedPermissions.Contains(apps.PermissionActAsUser) {
		var oAuthApp *model.OAuthApp
		oAuthApp, err = p.ensureOAuthApp(asAdmin, log, conf, app, trusted, in.ActingUserID)
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

	var message string
	if app.OnInstall != nil {
		resp := p.callApp(in, *app, apps.CallRequest{
			Call:    *app.OnInstall,
			Context: cc,
		})
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

	p.conf.Logger().Infof("Installed app.")

	p.dispatchRefreshBindingsEvent(in.ActingUserID)

	return app, message, nil
}

func (p *Proxy) ensureOAuthApp(client mmclient.Client, log utils.Logger, conf config.Config, app *apps.App, noUserConsent bool, actingUserID string) (*model.OAuthApp, error) {
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

func (p *Proxy) ensureBot(mm mmclient.Client, log utils.Logger, app *apps.App, icon io.Reader) error {
	bot := &model.Bot{
		Username:    strings.ToLower(string(app.AppID)),
		DisplayName: app.DisplayName,
		Description: fmt.Sprintf("Bot account for `%s` App.", app.DisplayName),
	}

	user, _ := mm.GetUserByUsername(bot.Username)
	if user == nil {
		err := mm.CreateBot(bot)
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
			bot, err = mm.EnableBot(user.Id)
			if err != nil {
				return err
			}
		}

		_, err := mm.GetBot(user.Id)
		if err != nil {
			err = mm.CreateBot(bot)
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

	if icon != nil {
		err := mm.SetProfileImage(app.BotUserID, icon)
		if err != nil {
			return errors.Wrap(err, "failed to update bot profile icon")
		}
	}

	// Create an access token on a fresh app install
	if app.RequestedPermissions.Contains(apps.PermissionActAsBot) &&
		app.BotAccessTokenID == "" {
		token, err := mm.CreateUserAccessToken(bot.UserId, "Mattermost App Token")
		if err != nil {
			return errors.Wrap(err, "failed to create bot user's access token")
		}

		app.BotAccessToken = token.Token
		app.BotAccessTokenID = token.Id
	}

	return nil
}

// getAppIcon gets the icon of a given app.
// Returns nil, nil if no app icon is defined in the manifest.
// The caller must close the returned io.ReadCloser if there is one.
func (p *Proxy) getAppIcon(app *apps.App) (io.ReadCloser, error) {
	iconPath := app.Manifest.Icon
	if iconPath == "" {
		return nil, nil
	}

	icon, status, err := p.getStatic(*app, iconPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get app icon")
	}

	if status != http.StatusOK {
		return nil, errors.Errorf("received %d status code while downloading bot icon for %v", status, app.Manifest.AppID)
	}

	return icon, nil
}
