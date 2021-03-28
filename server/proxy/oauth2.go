package proxy

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (p *Proxy) GetRemoteOAuth2RedirectURL(sessionID, actingUserID string, appID apps.AppID) (string, error) {
	app, err := p.store.App.Get(appID)
	if err != nil {
		return "", err
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return "", errors.Errorf("%s is not authorized to use OAuth2", appID)
	}

	creq := &apps.CallRequest{
		Call: *app.GetOAuth2RedirectURL.WithOverrides(apps.DefaultGetOAuth2RedirectURL),
		Type: apps.CallTypeSubmit,
		Context: p.conf.GetConfig().SetContextDefaultsForApp(appID,
			&apps.Context{
				ActingUserID: actingUserID,
			},
		),
	}
	cresp := p.Call(sessionID, actingUserID, creq)
	if cresp.Type == apps.CallResponseTypeError {
		return "", cresp
	}
	if cresp.Type != "" && cresp.Type != apps.CallResponseTypeOK {
		return "", errors.Errorf("oauth2: unexpected response type from the app: %q", cresp.Type)
	}
	redirectURL, ok := cresp.Data.(string)
	if !ok {
		return "", errors.Errorf("oauth2: unexpected data type from the app: %T, expected string (redirect URL)", cresp.Data)
	}

	return redirectURL, nil
}

func (p *Proxy) CompleteRemoteOAuth2(sessionID, actingUserID string, appID apps.AppID, urlValues map[string]interface{}) error {
	app, err := p.store.App.Get(appID)
	if err != nil {
		return err
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return errors.Errorf("%s is not authorized to use remote OAuth2", appID)
	}

	creq := &apps.CallRequest{
		Call: *app.OnOAuth2Complete.WithOverrides(apps.DefaultOnOAuth2Complete),
		Type: apps.CallTypeSubmit,
		Context: p.conf.GetConfig().SetContextDefaultsForApp(appID,
			&apps.Context{
				ActingUserID: actingUserID,
			},
		),
		Values: urlValues,
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
