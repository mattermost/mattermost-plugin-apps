package proxy

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (p *Proxy) GetRemoteOAuth2ConnectURL(in Incoming, appID apps.AppID) (string, error) {
	app, err := p.store.App.Get(appID)
	if err != nil {
		return "", err
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return "", errors.Errorf("%s is not authorized to use OAuth2", appID)
	}

	state, err := p.store.OAuth2.CreateState(in.ActingUserID)
	if err != nil {
		return "", err
	}

	cresp, _ := p.callApp(in, *app, apps.CallRequest{
		Call: app.GetOAuth2ConnectURL.WithDefault(apps.DefaultGetOAuth2ConnectURL),
		Values: map[string]interface{}{
			"state": state,
		},
	})
	if cresp.Type == apps.CallResponseTypeError {
		return "", &cresp
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

func (p *Proxy) CompleteRemoteOAuth2(in Incoming, appID apps.AppID, urlValues map[string]interface{}) error {
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
	err = p.store.OAuth2.ValidateStateOnce(urlState, in.ActingUserID)
	if err != nil {
		return err
	}

	cresp, _ := p.callApp(in, *app,
		apps.CallRequest{
			Call:   app.OnOAuth2Complete.WithDefault(apps.DefaultOnOAuth2Complete),
			Values: urlValues,
		})
	if cresp.Type == apps.CallResponseTypeError {
		return &cresp
	}
	if cresp.Type != "" && cresp.Type != apps.CallResponseTypeOK {
		return errors.Errorf("oauth2: unexpected response type from the app: %q", cresp.Type)
	}

	p.conf.Telemetry().TrackOAuthComplete(string(appID), in.ActingUserID)

	return nil
}
