package proxy

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// InvokeGetRemoteOAuth2ConnectURL returns the URL for the user to open, that would
// start the OAuth2 authentication flow. r.ActingUser and r.ToApp must be already set.
func (p *Proxy) InvokeGetRemoteOAuth2ConnectURL(r *incoming.Request) (string, error) {
	app, err := p.getEnabledDestination(r)
	if err != nil {
		return "", err
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return "", errors.Errorf("%s is not authorized to use OAuth2", app.AppID)
	}

	state, err := p.store.OAuth2.CreateState(r.ActingUserID())
	if err != nil {
		return "", err
	}

	call := app.GetOAuth2ConnectURL.WithDefault(apps.DefaultGetOAuth2ConnectURL)
	cresp := p.call(r, app, call, nil, "state", state)
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

func (p *Proxy) InvokeCompleteRemoteOAuth2(r *incoming.Request, urlValues map[string]interface{}) error {
	app, err := p.getEnabledDestination(r)
	if err != nil {
		return err
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return utils.NewUnauthorizedError("%s is not authorized to use remote OAuth2", app.AppID)
	}

	urlState, _ := urlValues["state"].(string)
	if urlState == "" {
		return utils.NewUnauthorizedError("no state arg in the URL")
	}
	err = p.store.OAuth2.ValidateStateOnce(urlState, r.ActingUserID())
	if err != nil {
		return err
	}

	cresp := p.callApp(r, app, apps.CallRequest{
		Call:    app.OnOAuth2Complete.WithDefault(apps.DefaultOnOAuth2Complete),
		Context: apps.Context{},
		Values:  urlValues,
	}, false)
	if cresp.Type == apps.CallResponseTypeError {
		return &cresp
	}
	if cresp.Type != "" && cresp.Type != apps.CallResponseTypeOK {
		return errors.Errorf("oauth2: unexpected response type from the app: %q", cresp.Type)
	}

	p.conf.Telemetry().TrackOAuthComplete(string(app.AppID), r.ActingUserID())

	return nil
}
