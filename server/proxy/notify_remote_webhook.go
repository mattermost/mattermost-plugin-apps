// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"crypto/subtle"
	"encoding/json"
	"net/url"
	"path"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	appspath "github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (p *Proxy) NotifyRemoteWebhook(appID apps.AppID, req httputils.ServerlessRequest) error {
	app, err := p.store.App.Get(appID)
	if err != nil {
		return err
	}
	if !p.appIsEnabled(*app) {
		return errors.Errorf("%s is disabled", app.AppID)
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteWebhooks) {
		return utils.NewForbiddenError("%s does not have permission %s", app.AppID, apps.PermissionRemoteWebhooks)
	}

	switch app.RemoteWebhookAuthType {
	case "", apps.NoAuth:

	case apps.SecretAuth:
		var q url.Values
		q, err = url.ParseQuery(req.RawQuery)
		if err != nil {
			return utils.NewForbiddenError(err)
		}
		secret := q.Get("secret")
		if secret == "" {
			return utils.NewInvalidError("webhook secret was not provided")
		}
		if subtle.ConstantTimeCompare([]byte(secret), []byte(app.WebhookSecret)) != 1 {
			return utils.NewInvalidError("webhook secret mismatched")
		}

	default:
		return errors.Errorf("%s is not a known webhook authentication type", app.RemoteWebhookAuthType)
	}

	up, err := p.upstreamForApp(*app)
	if err != nil {
		return err
	}

	var datav interface{}
	err = json.Unmarshal([]byte(req.Body), &datav)
	if err != nil {
		// if the data can not be decoded as JSON, send it "as is", as a string.
		datav = req.Body
	}

	conf := p.conf.Get()
	cc := contextForApp(*app, apps.Context{}, conf)
	// Set acting user to bot.
	cc.ActingUserID = app.BotUserID
	cc.ActingUserAccessToken = app.BotAccessToken

	// TODO: do we need to customize the Expand & State for the webhook Call?
	return upstream.Notify(up, *app, apps.CallRequest{
		Call: apps.Call{
			Path: path.Join(appspath.Webhook, req.Path),
		},
		Context: cc,
		Values: map[string]interface{}{
			"headers":    req.Headers,
			"data":       datav,
			"httpMethod": req.HTTPMethod,
			"rawQuery":   req.RawQuery,
		},
	})
}
