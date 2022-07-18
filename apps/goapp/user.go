package goapp

import (
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type User struct {
	MattermostID string
	RemoteID     string
	Token        *oauth2.Token
}

func (app *App) RemoveConnectedUser(creq CallRequest) error {
	asActingUser := appclient.AsActingUser(creq.Context)
	err := asActingUser.StoreOAuth2User(nil)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to removed the user record"))
	}

	creq.Log.Debugw("Removed user record", "id", creq.ActingUserID())
	return nil
}

func (app *App) StoreConnectedUser(creq CallRequest, user *User) error {
	if user == nil {
		return app.RemoveConnectedUser(creq)
	}

	asActingUser := appclient.AsActingUser(creq.Context)
	user.MattermostID = creq.ActingUserID()
	err := asActingUser.StoreOAuth2User(user)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to store the user record"))
	}

	accessTokenLog := ""
	expires := ""
	refreshTokenLog := ""
	if user.Token != nil {
		accessTokenLog = utils.LastN(user.Token.AccessToken, 4)
		expires = user.Token.Expiry.Format(time.RFC822)
		refreshTokenLog = utils.LastN(user.Token.RefreshToken, 4)
	}
	creq.Log.Debugw("Updated user record", "id", user.MattermostID, "access_token", accessTokenLog, "expires", expires, "refresh_token", refreshTokenLog)
	return nil
}

func OAuth2Logger(l utils.Logger, a *apps.OAuth2App, u *User) utils.Logger {
	if a != nil {
		l = l.With(
			"remote_url", a.RemoteRootURL,
			"client_id", a.ClientID,
			"client_secret", utils.LastN(a.ClientSecret, 4),
		)
		if a.Data != nil {
			l = l.With("app_data", a.Data)
		}
	}
	if u != nil {
		l = l.With(
			"user_id", u.MattermostID,
			"access_token", utils.LastN(u.Token.AccessToken, 4),
			"refresh_token", utils.LastN(u.Token.RefreshToken, 4),
			"token_expiry", u.Token.Expiry.String(),
		)
	}
	return l
}
