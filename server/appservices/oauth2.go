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

func (a *AppServices) StoreOAuth2App(r *incoming.Request, data []byte) error {
	if err := r.Check(
		r.RequireActingUser,
		r.RequireUserPermission(model.PermissionManageSystem),
		r.RequireSourceApp,
	); err != nil {
		return err
	}

	var oapp apps.OAuth2App
	err := json.Unmarshal(data, &oapp)
	if err != nil {
		return utils.NewInvalidError(errors.Wrap(err, "OAuth2App is not valid JSON"))
	}

	appID := r.SourceAppID()
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

	r.API.Mattermost.Frontend.PublishWebSocketEvent(config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{})

	return nil
}

func (a *AppServices) StoreOAuth2User(r *incoming.Request, data []byte) error {
	if err := r.Check(
		r.RequireActingUser,
		r.RequireSourceApp,
		r.RequireActingUserIsNotBot,
	); err != nil {
		return err
	}
	if !json.Valid(data) {
		return utils.NewInvalidError("payload is not valid JSON")
	}

	appID := r.SourceAppID()
	actingUserID := r.ActingUserID()
	app, err := a.store.App.Get(appID)
	if err != nil {
		return err
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return utils.NewUnauthorizedError("%s is not authorized to use remote OAuth2", app.AppID)
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

	r.API.Mattermost.Frontend.PublishWebSocketEvent(config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{UserId: actingUserID})
	return nil
}

// GetOAuth2User returns the stored OAuth2 user data for a given user and app.
// If err != nil, the returned data is always valid JSON.
func (a *AppServices) GetOAuth2User(r *incoming.Request) ([]byte, error) {
	if err := r.Check(
		r.RequireActingUser,
		r.RequireActingUserIsNotBot,
		r.RequireSourceApp,
	); err != nil {
		return nil, err
	}

	appID := r.SourceAppID()
	actingUserID := r.ActingUserID()
	app, err := a.store.App.Get(appID)
	if err != nil {
		return nil, err
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return nil, utils.NewUnauthorizedError("%s is not authorized to use remote OAuth2", app.AppID)
	}

	data, err := a.store.OAuth2.GetUser(appID, actingUserID)
	if err != nil && !errors.Is(err, utils.ErrNotFound) {
		return nil, err
	}

	return data, nil
}
