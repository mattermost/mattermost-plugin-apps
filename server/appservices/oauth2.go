package appservices

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

func (a *AppServices) CreateOAuth2State(actingUserID string) (string, error) {
	if err := a.ensureFromUser(actingUserID); err != nil {
		return "", err
	}
	return a.store.OAuth2.CreateState(actingUserID)
}

func (a *AppServices) StoreOAuth2App(botUserID string, oapp apps.OAuth2App) error {
	app, err := a.findBot(botUserID)
	if err != nil {
		return err
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return utils.NewUnauthorizedError("%s is not authorized to use remote OAuth2", app.AppID)
	}
	if err = a.ensureFromBot(botUserID); err != nil {
		return err
	}

	app.RemoteOAuth2 = oapp
	err = a.store.App.Save(app)
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
	return a.store.OAuth2.SaveUser(actingUserID, string(appID), ref)
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
	return a.store.OAuth2.GetUser(actingUserID, string(appID), ref)
}

func (a *AppServices) findBot(botUserID string) (*apps.App, error) {
	for _, app := range a.store.App.AsMap() {
		if app.BotUserID == botUserID {
			return app, nil
		}
	}
	return nil, utils.ErrNotFound
}
