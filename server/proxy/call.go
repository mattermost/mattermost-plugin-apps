// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (p *Proxy) Call(in Incoming, creq apps.CallRequest) apps.ProxyCallResponse {
	if creq.Context.AppID == "" {
		return apps.NewProxyCallResponse(
			apps.NewErrorCallResponse(
				utils.NewInvalidError("app_id is not set in Context, don't know what app to call")), nil)
	}

	app, err := p.store.App.Get(creq.Context.AppID)
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

	cc := creq.Context
	cc = in.updateContext(cc)
	cc, err = p.expandContext(in, app, &cc, creq.Expand)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	creq.Context = cc

	cresp := upstream.Call(up, *app, creq)
	if cresp.Type == "" {
		cresp.Type = apps.CallResponseTypeOK
	}

	if cresp.Form != nil {
		if cresp.Form.Icon != "" {
			conf, _, log := p.conf.Basic()
			log = log.With("app_id", app.AppID)
			icon, err := normalizeStaticPath(conf, cc.AppID, cresp.Form.Icon)
			if err != nil {
				log.WithError(err).Debugw("Invalid icon path in form. Ignoring it.", "icon", cresp.Form.Icon)
				cresp.Form.Icon = ""
			} else {
				cresp.Form.Icon = icon
			}
			clean, problems := cleanForm(*cresp.Form)
			for _, prob := range problems {
				log.WithError(prob).Debugw("invalid form")
			}
			cresp.Form = &clean
		}
	}

	return cresp
}
