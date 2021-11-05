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
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy/request"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// InstallApp installs an App.
//  - cc is the Context that will be passed down to the App's OnInstall callback.
func (p *Proxy) InstallApp(c *request.Context, cc apps.Context, appID apps.AppID, deployType apps.DeployType, trusted bool, secret string) (*apps.App, string, error) {
	conf := p.conf.Get()
	m, err := p.store.Manifest.Get(appID)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to find manifest to install app")
	}
	if !m.Contains(deployType) {
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

	if app.GrantedPermissions.Contains(apps.PermissionRemoteWebhooks) &&
		app.RemoteWebhookAuthType == apps.SecretAuth || app.RemoteWebhookAuthType == "" {
		app.WebhookSecret = model.NewId()
	}

	icon, err := p.getAppIcon(c, *app)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed get bot icon")
	}
	if icon != nil {
		defer icon.Close()
	}

	// See if the app is inaaccessible. Call its ping path with nothing
	// expanded, ignore 404 errors coming back and consider everything else a
	// "success".
	//
	// Note that this check is often ineffective, but "the best we can do"
	// before we start the diffcult-to-revert install process.
	_, err = p.callApp(c, *app, apps.CallRequest{
		Call:    apps.DefaultPing,
		Context: cc,
	})
	if err != nil && errors.Cause(err) != utils.ErrNotFound {
		return nil, "", errors.Wrapf(err, "failed to install, %s path is not accessible", apps.DefaultPing.Path)
	}

	err = p.ensureBot(c, app, icon)
	if err != nil {
		return nil, "", err
	}

	if app.GrantedPermissions.Contains(apps.PermissionActAsUser) {
		var oAuthApp *model.OAuthApp
		oAuthApp, err = p.ensureOAuthApp(c, conf, *app, trusted, c.ActingUserID())
		if err != nil {
			return nil, "", err
		}
		app.MattermostOAuth2 = oAuthApp
	}

	err = p.store.App.Save(*app)
	if err != nil {
		return nil, "", err
	}

	message := fmt.Sprintf("Installed %s.", app.DisplayName)
	if app.OnInstall != nil {
		cresp := p.call(c, *app, *app.OnInstall, &cc)
		if cresp.Type == apps.CallResponseTypeError {
			// TODO: should fail and roll back.
			c.Log.WithError(cresp).Warnf("Installed %s, despite on_install failure.", app.AppID)
			message = fmt.Sprintf("Installed %s, despite on_install failure: %s", app.AppID, cresp.Error())
		} else if cresp.Markdown != "" {
			message += "\n\n" + cresp.Markdown
		}
	} else if len(app.GrantedLocations) > 0 {
		// Make sure the app's binding call is accessible.
		cresp := p.call(c, *app, app.Bindings.WithDefault(apps.DefaultBindings), &cc)
		if cresp.Type == apps.CallResponseTypeError {
			// TODO: should fail and roll back.
			c.Log.WithError(cresp).Warnf("Installed %s, despite bindings failure.", app.AppID)
			message = fmt.Sprintf("Installed %s despite bindings failure: %s", app.AppID, cresp.Error())
		}
	}

	p.conf.Telemetry().TrackInstall(string(app.AppID), string(app.DeployType))

	p.dispatchRefreshBindingsEvent(c.ActingUserID())

	c.Log.Infof(message)

	return app, message, nil
}

func (p *Proxy) ensureOAuthApp(c *request.Context, conf config.Config, app apps.App, noUserConsent bool, actingUserID string) (*model.OAuthApp, error) {
	mm := p.conf.MattermostAPI()
	if app.MattermostOAuth2 != nil {
		c.Log.Debugw("App install flow: Using existing OAuth2 App", "id", app.MattermostOAuth2.Id)

		return app.MattermostOAuth2, nil
	}

	oauth2CallbackURL := conf.AppURL(app.AppID) + path.MattermostOAuth2Complete

	oauthApp := &model.OAuthApp{
		CreatorId:          actingUserID,
		Name:               app.DisplayName,
		Description:        app.Description,
		CallbackUrls:       []string{oauth2CallbackURL},
		Homepage:           app.HomepageURL,
		IsTrusted:          noUserConsent,
		Scopes:             nil,
		AppsFrameworkAppID: string(app.AppID),
	}
	err := mm.OAuth.Create(oauthApp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create OAuth2 App")
	}

	c.Log.Debugw("App install flow: Created OAuth2 App", "id", oauthApp.Id)

	return oauthApp, nil
}

func (p *Proxy) ensureBot(c *request.Context, app *apps.App, icon io.Reader) error {
	mm := p.conf.MattermostAPI()
	bot := &model.Bot{
		Username:    strings.ToLower(string(app.AppID)),
		DisplayName: app.DisplayName,
		Description: fmt.Sprintf("Bot account for `%s` App.", app.DisplayName),
	}

	user, _ := mm.User.GetByUsername(bot.Username)
	if user == nil {
		err := mm.Bot.Create(bot)
		if err != nil {
			return err
		}

		c.Log.Debugw("App install flow: Created Bot Account ",
			"username", bot.Username)
	} else {
		if !user.IsBot {
			return errors.New("a user already owns the bot username")
		}

		// Check if disabled
		if user.DeleteAt != 0 {
			var err error
			bot, err = mm.Bot.UpdateActive(user.Id, true)
			if err != nil {
				return err
			}
		}

		_, err := mm.Bot.Get(user.Id, false)
		if err != nil {
			err = mm.Bot.Create(bot)
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
		err := mm.User.SetProfileImage(app.BotUserID, icon)
		if err != nil {
			return errors.Wrap(err, "failed to update bot profile icon")
		}
	}

	// Create an access token on a fresh app install
	if app.RequestedPermissions.Contains(apps.PermissionActAsBot) &&
		app.BotAccessTokenID == "" {
		// Use the Plugin API as OAuth sessions can't create access tokens
		token, err := p.conf.MattermostAPI().User.CreateAccessToken(bot.UserId, "Mattermost App Token")
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
func (p *Proxy) getAppIcon(c *request.Context, app apps.App) (io.ReadCloser, error) {
	iconPath := app.Manifest.Icon
	if iconPath == "" {
		return nil, nil
	}

	icon, status, err := p.getStatic(c, app, iconPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get app icon")
	}

	if status != http.StatusOK {
		return nil, errors.Errorf("received %d status code while downloading bot icon for %v",
			status, app.Manifest.AppID)
	}

	return icon, nil
}
