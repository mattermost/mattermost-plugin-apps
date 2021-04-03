package proxy

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

func (p *Proxy) GetRemoteOAuth2ConnectURL(sessionID, actingUserID string, appID apps.AppID) (string, error) {
	app, err := p.store.App.Get(appID)
	if err != nil {
		return "", err
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return "", errors.Errorf("%s is not authorized to use OAuth2", appID)
	}

	state, err := p.store.OAuth2.CreateState(actingUserID)
	if err != nil {
		return "", err
	}

	creq := &apps.CallRequest{
		Call: *app.GetOAuth2ConnectURL.WithOverrides(apps.DefaultGetOAuth2ConnectURL),
		Type: apps.CallTypeSubmit,
		Context: p.conf.GetConfig().SetContextDefaultsForApp(appID,
			&apps.Context{
				ActingUserID: actingUserID,
			},
		),
		Values: map[string]interface{}{
			"state": state,
		},
	}
	cresp := p.Call(sessionID, actingUserID, creq)
	if cresp.Type == apps.CallResponseTypeError {
		return "", cresp
	}
	if cresp.Type != "" && cresp.Type != apps.CallResponseTypeOK {
		return "", errors.Errorf("oauth2: unexpected response type from the app: %q", cresp.Type)
	}
	connectURL, ok := cresp.Data.(string)
	if !ok {
		return "", errors.Errorf("oauth2: unexpected data type from the app: %T, expected string (connect URL)", cresp.Data)
	}

	return connectURL, nil
}

func (p *Proxy) CompleteRemoteOAuth2(sessionID, actingUserID string, appID apps.AppID, urlValues map[string]interface{}) error {
	app, err := p.store.App.Get(appID)
	if err != nil {
		return err
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return utils.NewUnauthorizedError("%s is not authorized to use remote OAuth2", appID)
	}

	urlState, _ := urlValues["state"].(string)
	if urlState == "" {
		return utils.NewUnauthorizedError("no state arg in the URL")
	}
	err = p.store.OAuth2.ValidateStateOnce(urlState, actingUserID)
	if err != nil {
		return err
	}

	creq := &apps.CallRequest{
		Call:    *app.OnOAuth2Complete.WithOverrides(apps.DefaultOnOAuth2Complete),
		Type:    apps.CallTypeSubmit,
		Context: p.conf.GetConfig().SetContextDefaultsForApp(appID, &apps.Context{}),
		Values:  urlValues,
	}
	cresp := p.Call(sessionID, actingUserID, creq)
	if cresp.Type == apps.CallResponseTypeError {
		return cresp
	}
	if cresp.Type != "" && cresp.Type != apps.CallResponseTypeOK {
		return errors.Errorf("oauth2: unexpected response type from the app: %q", cresp.Type)
	}

	return nil
}
