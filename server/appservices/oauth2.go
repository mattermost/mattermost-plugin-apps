package appservices

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (a *AppServices) StoreOAuth2App(r *incoming.Request, appID apps.AppID, actingUserID string, data []byte) error {
	var oapp apps.OAuth2App
	err := json.Unmarshal(data, &oapp)
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

	oldData, _ := json.Marshal(app.RemoteOAuth2)
	if bytes.Equal(oldData, data) {
		return nil
	}

	app.RemoteOAuth2 = oapp
	err = a.store.App.Save(r, *app)
	if err != nil {
		return err
	}

	a.conf.MattermostAPI().Frontend.PublishWebSocketEvent(config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{})

	return nil
}

func (a *AppServices) StoreOAuth2User(r *incoming.Request, appID apps.AppID, actingUserID string, data []byte) error {
	if !json.Valid(data) {
		return utils.NewInvalidError("payload is no valid json")
	}

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

	oldData, err := a.store.OAuth2.GetUser(appID, actingUserID)
	if err != nil {
		return err
	}
	if bytes.Equal(data, oldData) {
		return nil
	}

	err = a.store.OAuth2.SaveUser(appID, actingUserID, data)
	if err != nil {
		return err
	}

	a.conf.MattermostAPI().Frontend.PublishWebSocketEvent(config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{UserId: actingUserID})
	return nil
}

// GetOAuth2User returns the stored OAuth2 user data for a given user and app.
// If err != nil, the returned data is always valid JSON.
func (a *AppServices) GetOAuth2User(_ *incoming.Request, appID apps.AppID, actingUserID string) ([]byte, error) {
	app, err := a.store.App.Get(appID)
	if err != nil {
		return nil, err
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return nil, utils.NewUnauthorizedError("%s is not authorized to use remote OAuth2", app.AppID)
	}

	if err = a.ensureFromUser(actingUserID); err != nil {
		return nil, err
	}

	data, err := a.store.OAuth2.GetUser(appID, actingUserID)
	if err != nil && !errors.Is(err, utils.ErrNotFound) {
		return nil, err
	}

	return data, nil
}
