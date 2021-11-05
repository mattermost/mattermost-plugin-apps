package appservices

import (
	"bytes"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (a *AppServices) StoreOAuth2App(appID apps.AppID, actingUserID string, oapp apps.OAuth2App) error {
	app, err := a.store.App.Get(appID)
	if err != nil {
		return err
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return utils.NewUnauthorizedError("%s is not authorized to use remote OAuth2", app.AppID)
	}

	app.RemoteOAuth2 = oapp
	err = a.store.App.Save(*app)
	if err != nil {
		return err
	}

	return nil
}

func (a *AppServices) StoreOAuth2User(appID apps.AppID, actingUserID string, ref []byte) error {
	app, err := a.store.App.Get(appID)
	if err != nil {
		return err
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return utils.NewUnauthorizedError("%s is not authorized to use remote OAuth2", app.AppID)
	}
	if err = a.ensureFromUser(actingUserID); err != nil {
		return err
	}

	var oauth2user []byte
	err = a.store.OAuth2.GetUser(app.BotUserID, actingUserID, &oauth2user)
	if err != nil {
		return err
	}

	// Trigger a bindings refresh if the OAuth2 user was updated
	if !bytes.Equal(ref, oauth2user) {
		a.conf.MattermostAPI().Frontend.PublishWebSocketEvent(config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{UserId: actingUserID})
	}

	return a.store.OAuth2.SaveUser(app.BotUserID, actingUserID, ref)
}

func (a *AppServices) GetOAuth2User(appID apps.AppID, actingUserID string, ref interface{}) error {
	app, err := a.store.App.Get(appID)
	if err != nil {
		return err
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return utils.NewUnauthorizedError("%s is not authorized to use remote OAuth2", app.AppID)
	}
	if err = a.ensureFromUser(actingUserID); err != nil {
		return err
	}
	return a.store.OAuth2.GetUser(app.BotUserID, actingUserID, ref)
}
