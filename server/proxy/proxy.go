// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (p *Proxy) Call(in Incoming, appID apps.AppID, creq apps.CallRequest) apps.ProxyCallResponse {
	app, err := p.store.App.Get(appID)
	if err != nil {
		return apps.NewProxyCallResponse(apps.NewErrorCallResponse(err), nil)
	}

	metadata := &apps.AppMetadataForClient{
		BotUserID:   app.BotUserID,
		BotUsername: app.BotUsername,
	}

	cresp := p.callApp(in, app, creq)
	return apps.NewProxyCallResponse(cresp, metadata)
}

func (p *Proxy) callApp(in Incoming, app *apps.App, creq apps.CallRequest) apps.CallResponse {
	if !p.appIsEnabled(app) {
		return apps.NewErrorCallResponse(errors.Errorf("%s is disabled", app.AppID))
	}

	if creq.Path[0] != '/' {
		return apps.NewErrorCallResponse(utils.NewInvalidError("call path must start with a %q: %q", "/", creq.Path))
	}
	cleanPath, err := utils.CleanPath(creq.Path)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	creq.Path = cleanPath

	up, err := p.upstreamForApp(app)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}

	cc := in.updateContext(creq.Context)
	creq.Context, err = p.expandContext(&cc, app, creq.Expand)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}

	callResponse := upstream.Call(up, *app, creq)
	if callResponse.Type == "" {
		callResponse.Type = apps.CallResponseTypeOK
	}

	if callResponse.Form != nil && callResponse.Form.Icon != "" {
		conf := p.conf.GetConfig()
		icon, err := normalizeStaticPath(conf, app.AppID, callResponse.Form.Icon)
		if err != nil {
			p.log.WithError(err).Debugw("Invalid icon path in form. Ignoring it.",
				"app_id", app.AppID,
				"icon", callResponse.Form.Icon)
			callResponse.Form.Icon = ""
		} else {
			callResponse.Form.Icon = icon
		}
	}

	return callResponse
}

func (p *Proxy) simpleCall(in Incoming, app *apps.App, call apps.Call) apps.CallResponse {
	return p.callApp(in, app, apps.CallRequest{Call: call})
}

// normalizeStaticPath converts a given URL to a absolute one pointing to a static asset if needed.
// If icon is an absolute URL, it's not changed.
// Otherwise assume it's a path to a static asset and the static path URL prepended.
func normalizeStaticPath(conf config.Config, appID apps.AppID, icon string) (string, error) {
	if !strings.HasPrefix(icon, "http://") && !strings.HasPrefix(icon, "https://") {
		cleanIcon, err := utils.CleanStaticPath(icon)
		if err != nil {
			return "", errors.Wrap(err, "invalid icon path")
		}

		icon = conf.StaticURL(appID, cleanIcon)
	}

	return icon, nil
}

func (p *Proxy) Notify(base apps.Context, subj apps.Subject) error {
	subs, err := p.store.Subscription.Get(subj, base.TeamID, base.ChannelID)
	if err != nil {
		return err
	}

	notify := func(sub apps.Subscription) error {
		call := sub.Call
		if call == nil {
			return errors.New("nothing to call")
		}

		creq := apps.CallRequest{Call: *call}
		app, err := p.store.App.Get(sub.AppID)
		if err != nil {
			return err
		}
		if !p.appIsEnabled(app) {
			return errors.Errorf("%s is disabled", app.AppID)
		}

		creq.Context, err = p.expandContext(&base, app, creq.Call.Expand)
		if err != nil {
			return err
		}
		creq.Context.Subject = subj

		up, err := p.upstreamForApp(app)
		if err != nil {
			return err
		}
		return upstream.Notify(up, *app, creq)
	}

	for _, sub := range subs {
		err := notify(sub)
		if err != nil {
			// TODO log err
			continue
		}
	}
	return nil
}

func (p *Proxy) NotifyRemoteWebhook(app apps.App, data []byte, webhookPath string) error {
	if !p.appIsEnabled(&app) {
		return errors.Errorf("%s is disabled", app.AppID)
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteWebhooks) {
		return utils.NewForbiddenError("%s does not have permission %s", app.AppID, apps.PermissionRemoteWebhooks)
	}

	up, err := p.upstreamForApp(&app)
	if err != nil {
		return err
	}

	var datav interface{}
	err = json.Unmarshal(data, &datav)
	if err != nil {
		// if the data can not be decoded as JSON, send it "as is", as a string.
		datav = string(data)
	}

	conf := p.conf.GetConfig()
	cc := forApp(&app, apps.Context{}, conf)
	cc.ActingUserID = app.BotUserID
	cc.ActingUserAccessToken = app.BotAccessToken

	// TODO: do we need to customize the Expand & State for the webhook Call?
	return upstream.Notify(up, app, apps.CallRequest{
		Call: apps.Call{
			Path: path.Join(apps.PathWebhook, webhookPath),
		},
		Context: cc,
		Values: map[string]interface{}{
			"data": datav,
		},
	})
}

func (p *Proxy) GetStatic(appID apps.AppID, path string) (io.ReadCloser, int, error) {
	app, err := p.store.App.Get(appID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, utils.ErrNotFound) {
			status = http.StatusNotFound
		}
		return nil, status, err
	}

	return p.getStatic(app, path)
}

func (p *Proxy) getStatic(app *apps.App, path string) (io.ReadCloser, int, error) {
	up, err := p.upstreamForApp(app)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return up.GetStatic(*app, path)
}

func (p *Proxy) upstreamForApp(app *apps.App) (upstream.Upstream, error) {
	if app.DeployType == apps.DeployBuiltin {
		u, ok := p.builtinUpstreams[app.AppID]
		if !ok {
			return nil, errors.Wrapf(utils.ErrNotFound, "no builtin %s", app.AppID)
		}
		return u, nil
	}

	err := CanDeploy(p, app.DeployType)
	if err != nil {
		return nil, err
	}

	upv, ok := p.upstreams.Load(app.DeployType)
	if !ok {
		return nil, utils.NewInvalidError("invalid or unsupported upstream type: %s", app.DeployType)
	}
	up, ok := upv.(upstream.Upstream)
	if !ok {
		return nil, utils.NewInvalidError("invalid Upstream for: %s", app.DeployType)
	}
	return up, nil
}
