package appservices

import (
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (a *AppServices) StoreOAuth2App(appID apps.AppID, actingUserID string, oapp apps.OAuth2App) error {
	err := utils.EnsureSysAdmin(a.conf.MattermostAPI(), actingUserID)
	if err != nil {
		return err
	}

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

func (a *AppServices) StoreOAuth2User(appID apps.AppID, actingUserID string, ref interface{}) error {
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
	var oauth2user interface{}
	_ = a.GetOAuth2User(appID, actingUserID, &oauth2user)

	var refinterface interface{}
	json.Unmarshal(ref.([]byte), &refinterface)
	eq := reflect.DeepEqual(refinterface, oauth2user)
	if !eq {
		a.mm.Frontend.PublishWebSocketEvent(config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{UserId: actingUserID})
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
