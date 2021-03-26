package proxy

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (p *Proxy) GetOAuth2RedirectURL(appID apps.AppID, actingUserID, token string) (string, error) {
	app, err := p.store.App.Get(appID)
	if err != nil {
		return "", err
	}

	cresp := p.Call(apps.SessionToken(token), &apps.CallRequest{
		Call: *app.OnOAuth2Redirect.WithOverrides(apps.DefaultOnOAuth2RedirectCall),
		Type: apps.CallTypeSubmit,
		Context: p.conf.GetConfig().SetContextDefaultsForApp(appID,
			&apps.Context{
				ActingUserID: actingUserID,
			},
		),
	})
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

func (p *Proxy) CompleteOAuth2(appID apps.AppID, actingUserID, token string, urlValues map[string]interface{}) error {
	app, err := p.store.App.Get(appID)
	if err != nil {
		return err
	}

	cresp := p.Call(apps.SessionToken(token), &apps.CallRequest{
		Call: *app.OnOAuth2Complete.WithOverrides(apps.DefaultOnOAuth2CompleteCall),
		Type: apps.CallTypeSubmit,
		Context: p.conf.GetConfig().SetContextDefaultsForApp(appID,
			&apps.Context{
				ActingUserID: actingUserID,
			},
		),
		Values: urlValues,
	})
	if cresp.Type == apps.CallResponseTypeError {
		return cresp
	}
	if cresp.Type != "" && cresp.Type != apps.CallResponseTypeOK {
		return errors.Errorf("oauth2: unexpected response type from the app: %q", cresp.Type)
	}

	return nil
}
