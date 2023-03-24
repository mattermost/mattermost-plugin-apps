// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// InstallApp installs an App.
//   - cc is the Context that will be passed down to the App's OnInstall callback.
func (p *Proxy) InstallApp(r *incoming.Request, cc apps.Context, appID apps.AppID, deployType apps.DeployType, trusted bool, secret string) (app *apps.App, message string, err error) {
	prevlog := r.Log
	start := time.Now()
	r.Log = r.Log.With("app_id", appID, "deploy_type", deployType)
	defer func() {
		log := r.Log.With("elapsed", time.Since(start).String())
		r.Log = prevlog
		if err != nil {
			log.WithError(err).Errorf("InstallApp failed")
		} else {
			r.Log.Infof(message)
		}
	}()

	if err = r.Check(
		r.RequireActingUser,
		r.RequireSysadminOrPlugin,
	); err != nil {
		return nil, "", err
	}

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

	r.Log.Debugf("app install flow: %s can be installed, proceeding", appID)
	app, err = p.store.App.Get(appID)
	if err != nil {
		if !errors.Is(err, utils.ErrNotFound) {
			return nil, "", errors.Wrap(err, "failed looking for existing app")
		}
		r.Log.Debugf("app install flow: no previous installation detected")
		app = &apps.App{}
	} else {
		r.Log.Debugf("app install flow: previous installation detected")
	}

	app.DeployType = deployType
	app.Manifest = *m
	if app.Disabled {
		r.Log.Debugf("app install flow: re-enabling previously disabled app")
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
	r.Log.Debugf("app install flow: updated app configuration")

	err = p.pingApp(r.Ctx(), app)
	if err != nil {
		return nil, "", errors.Wrapf(err, "failed to install, %s path is not accessible", apps.DefaultPing.Path)
	}
	r.Log.Debugf("app install flow: app is pingable")

	icon, err := p.getAppIcon(r, app)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to get bot icon")
	}
	if icon != nil {
		defer icon.Close()
	}
	r.Log.Debugf("app install flow: downloaded the app's icon")

	err = p.ensureBot(r, app, icon)
	if err != nil {
		return nil, "", err
	}
	r.Log.Debugf("app install flow: configured bot account for the app")

	err = p.ensureOAuthApp(r, conf, app, trusted, r.ActingUserID())
	if err != nil {
		return nil, "", err
	}
	r.Log.Debugf("app install flow: configured OAuth2 account for the app")

	err = p.store.App.Save(r, *app)
	if err != nil {
		return nil, "", err
	}
	r.Log.Debugf("app install flow: stored updated app configuration")

	message = fmt.Sprintf("Installed app `%s`: %s.", app.AppID, app.DisplayName)
	if app.OnInstall != nil {
		cresp := p.call(r, app, *app.OnInstall, &cc)
		if cresp.Type == apps.CallResponseTypeError {
			// TODO: should fail and roll back.
			r.Log.WithError(cresp).Warnf("%s, despite on_install failure.", message)
			message += fmt.Sprintf(", despite on_install failure: %s", cresp.Error())
		} else if cresp.Text != "" {
			message += "\n\n" + cresp.Text
		}
	} else if len(app.GrantedLocations) > 0 {
		// Make sure the app's binding call is accessible.
		cresp := p.call(r, app, app.Bindings.WithDefault(apps.DefaultBindings), &cc)
		if cresp.Type == apps.CallResponseTypeError {
			// TODO: should fail and roll back.
			r.Log.WithError(cresp).Warnf("%s, despite bindings failure.", message)
			message += fmt.Sprintf(", despite bindings failure: %s", cresp.Error())
		}
	}

	p.conf.Telemetry().TrackInstall(string(app.AppID), string(app.DeployType))

	p.dispatchRefreshBindingsEvent(r.ActingUserID())

	return app, message, nil
}

func (p *Proxy) ensureOAuthApp(r *incoming.Request, conf config.Config, app *apps.App, noUserConsent bool, actingUserID string) error {
	mm := p.conf.MattermostAPI()
	if app.MattermostOAuth2 != nil {
		r.Log.Debugw("app install flow: Using existing OAuth2 App", "id", app.MattermostOAuth2.Id)
		return nil
	}

	oauth2CallbackURL := conf.AppURL(app.AppID) + path.MattermostOAuth2Complete

	oAuthApp := &model.OAuthApp{
		CreatorId:       actingUserID,
		Name:            app.DisplayName,
		Description:     app.Description,
		CallbackUrls:    []string{oauth2CallbackURL},
		Homepage:        app.HomepageURL,
		IsTrusted:       noUserConsent,
		MattermostAppID: string(app.AppID),
	}
	err := mm.OAuth.Create(oAuthApp)
	if err != nil {
		return errors.Wrap(err, "failed to create OAuth2 App")
	}

	app.MattermostOAuth2 = oAuthApp

	r.Log.Debugw("app install flow: created OAuth2 App", "id", app.MattermostOAuth2.Id)

	return nil
}

func (p *Proxy) createAndValidateBot(r *incoming.Request, bot *model.Bot) error {
	mm := p.conf.MattermostAPI()

	err := mm.Bot.Create(bot)
	if err != nil {
		return err
	}

	r.Log.Debugw("app install flow: created Bot Account ",
		"username", bot.Username, "bot_user_id", bot.UserId)

	const tryCount = 5
	delay := 50 * time.Millisecond
	i := 0
	for {
		bb, ee := mm.Bot.Get(bot.UserId, false)
		if bb != nil {
			r.Log.Debugw("app install flow: validated Bot Account", "bot_user_id", bb.UserId)
			return nil
		}

		i++
		if i >= tryCount {
			r.Log.Debugw("app install flow: failed to validate Bot Account, try re-installing the App", "try", i+1, "error", ee)
			return ee
		}
		r.Log.Debugw("app install flow: retrying Bot Account after delay", "try", i+1, "error", ee, "delay", delay)
		time.Sleep(delay)
		delay *= 2
	}
}

func (p *Proxy) ensureBot(r *incoming.Request, app *apps.App, icon io.Reader) error {
	username := strings.ToLower(string(app.AppID))
	err := p.ensureBotNamed(r, app, icon, username)
	if errors.Cause(err) == utils.ErrAlreadyExists {
		r.Log.Debugf("app install flow: a user already owns the username %s, trying %s-app", username, username)
		return p.ensureBotNamed(r, app, icon, username+"-app")
	}
	return err
}

func (p *Proxy) ensureBotNamed(r *incoming.Request, app *apps.App, icon io.Reader, username string) error {
	mm := p.conf.MattermostAPI()
	bot := &model.Bot{
		Username:    username,
		DisplayName: app.DisplayName,
		Description: fmt.Sprintf("Bot account for `%s` App.", app.DisplayName),
	}

	user, _ := mm.User.GetByUsername(bot.Username)
	if user == nil {
		r.Log.Debugw("app install flow: using existing Bot Account", "username", bot.Username, "id", bot.UserId)
		err := p.createAndValidateBot(r, bot)
		if err != nil {
			return err
		}
	} else {
		if !user.IsBot {
			r.Log.Debugf("app install flow: user %s exists and is not a bot", bot.Username)
			return utils.NewAlreadyExistsError("a user already owns the bot username %s", bot.Username)
		}

		// Check if disabled
		if user.DeleteAt != 0 {
			var err error
			r.Log.Debugf("app install flow: bot user %s was deleted, reactivating", bot.Username)
			bot, err = mm.Bot.UpdateActive(user.Id, true)
			if err != nil {
				return err
			}
		}

		// include deleted bots in the request, it appears that UpdateActive
		// does not propagate instantly. The userID and username remain the
		// same.
		_, err := mm.Bot.Get(user.Id, true)
		if err != nil {
			r.Log.Debugf("app install flow: failed to get the bot record for user %s (%s): %v; attempting to create a new one...", bot.UserId, bot.Username, err)
			err = p.createAndValidateBot(r, bot)
			if err != nil {
				return err
			}
		} else {
			bot.UserId = user.Id
			bot.Username = user.Username

			if user.GetDisplayName(model.ShowFullName) != app.DisplayName {
				// The app has changed its display name
				bot.DisplayName = app.DisplayName
				bot.Description = fmt.Sprintf("Bot account for `%s` App.", app.DisplayName)

				patchData := model.BotPatch{
					Username:    &bot.Username,
					DisplayName: &bot.DisplayName,
					Description: &bot.Description,
				}

				r.Log.Debugw("app install flow: patching existing Bot Account", "username", bot.Username, "id", bot.UserId)
				mm.Bot.Patch(bot.UserId, &patchData)
			}

			r.Log.Debugw("app install flow: using existing Bot Account", "username", bot.Username, "id", bot.UserId)
		}
	}
	app.BotUserID = bot.UserId
	app.BotUsername = bot.Username

	if icon != nil {
		err := mm.User.SetProfileImage(app.BotUserID, icon)
		if err != nil {
			r.Log.WithError(err).Errorf("app install flow: failed to update Bot Account profile icon, try re-installing")
		}
	}

	return nil
}

// getAppIcon gets the icon of a given app.
// Returns nil, nil if no app icon is defined in the manifest.
// The caller must close the returned io.ReadCloser if there is one.
func (p *Proxy) getAppIcon(r *incoming.Request, app *apps.App) (io.ReadCloser, error) {
	iconPath := app.Manifest.Icon
	if iconPath == "" {
		return nil, nil
	}

	icon, status, err := p.getStatic(r, app, iconPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to download icon: %s", iconPath)
	}

	if status != http.StatusOK {
		return nil, errors.Errorf("failed to download icon: %s: received status code %v", iconPath, status)
	}

	return icon, nil
}
