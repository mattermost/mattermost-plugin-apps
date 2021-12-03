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

func (a *AppServices) StoreOAuth2App(r *incoming.Request, appID apps.AppID, actingUserID string, oapp apps.OAuth2App) error {
	app, err := a.store.App.Get(r, appID)
	if err != nil {
		return err
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return utils.NewUnauthorizedError("%s is not authorized to use remote OAuth2", app.AppID)
	}

	app.RemoteOAuth2 = oapp
	err = a.store.App.Save(r, *app)
	if err != nil {
		return err
	}

	return nil
}

func (a *AppServices) StoreOAuth2User(r *incoming.Request, appID apps.AppID, actingUserID string, data []byte) error {
	if !json.Valid(data) {
		return utils.NewInvalidError("payload is no valid json")
	}

	app, err := a.store.App.Get(r, appID)
	if err != nil {
		return err
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return utils.NewUnauthorizedError("%s is not authorized to use remote OAuth2", app.AppID)
	}

	if err = a.ensureFromUser(actingUserID); err != nil {
		return err
	}

	oldData, err := a.store.OAuth2.GetUser(r, appID, actingUserID)
	if err != nil {
		return err
	}

	// Trigger a bindings refresh if the OAuth2 user was updated
	if !bytes.Equal(data, oldData) {
		a.conf.MattermostAPI().Frontend.PublishWebSocketEvent(config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{UserId: actingUserID})
	}

	return a.store.OAuth2.SaveUser(r, appID, actingUserID, data)
}

// GetOAuth2User returns the stored OAuth2 user data for a given user and app.
// If err != nil, the returned data is always valid JSON.
func (a *AppServices) GetOAuth2User(r *incoming.Request, appID apps.AppID, actingUserID string) ([]byte, error) {
	app, err := a.store.App.Get(r, appID)
	if err != nil {
		return nil, err
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return nil, utils.NewUnauthorizedError("%s is not authorized to use remote OAuth2", app.AppID)
	}

	if err = a.ensureFromUser(actingUserID); err != nil {
		return nil, err
	}

	data, err := a.store.OAuth2.GetUser(r, appID, actingUserID)
	if err != nil && !errors.Is(err, utils.ErrNotFound) {
		return nil, err
	}

	if len(data) == 0 {
		// Ensure valid json is returned even if no data is set yet
		data = []byte(string("{}"))
	}

	return data, nil
}
