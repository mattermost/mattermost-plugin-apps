package appservices

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

func (a *AppServices) StoreOAuth2App(appID apps.AppID, actingUserID string, oapp apps.OAuth2App) error {
	err := utils.EnsureSysAdmin(a.mm, actingUserID)
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
	return a.store.OAuth2.SaveUser(appID, actingUserID, ref)
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
	return a.store.OAuth2.GetUser(appID, actingUserID, ref)
}
