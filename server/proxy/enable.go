// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

// EnableApp is never called. Context vetting needs to be done by the caller via p.CleanUserCallContext, though we're temporarily doing it here instead.
func (p *Proxy) EnableApp(cc *apps.Context, app *apps.App) (md.MD, error) {
	err := utils.EnsureSysAdmin(p.mm, cc.ActingUserID)
	if err != nil {
		return "", err
	}

	if !app.Disabled {
		return "no change.", nil
	}

	cc, err = p.CleanUserCallContext(cc.ActingUserID, cc)
	if err != nil {
		return "", err
	}

	app.Disabled = false
	err = p.store.App.Save(app)
	if err != nil {
		return "", err
	}

	var message md.MD
	if app.OnEnable != nil {
		resp := p.Call("", cc.ActingUserID, &apps.CallRequest{
			Call:    *app.OnEnable,
			Context: cc,
		})
		if resp.Type == apps.CallResponseTypeError {
			p.mm.Log.Warn("OnEnable failed, app enabled anyway", "err", resp.Error(), "app_id", app.AppID)
		}
		message = resp.Markdown
	}

	p.mm.Frontend.PublishWebSocketEvent(config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{UserId: cc.ActingUserID})

	return md.Markdownf("%s is now enabled:\n%s", app.DisplayName, message), nil
}

func (p *Proxy) DisableApp(cc *apps.Context, app *apps.App) (md.MD, error) {
	err := utils.EnsureSysAdmin(p.mm, cc.ActingUserID)
	if err != nil {
		return "", err
	}

	if app.Disabled {
		return "no change.", nil
	}

	cc, err = p.CleanUserCallContext(cc.ActingUserID, cc)
	if err != nil {
		return "", err
	}

	app.Disabled = true

	var message md.MD
	if app.OnDisable != nil {
		resp := p.Call("", cc.ActingUserID, &apps.CallRequest{
			Call:    *app.OnDisable,
			Context: cc,
		})
		if resp.Type == apps.CallResponseTypeError {
			p.mm.Log.Warn("OnDisable failed, app disabled anyway", "err", resp.Error(), "app_id", app.AppID)
		}
		message = resp.Markdown
	}

	err = p.store.App.Save(app)
	if err != nil {
		return "", err
	}

	p.mm.Frontend.PublishWebSocketEvent(config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{UserId: cc.ActingUserID})

	return md.Markdownf("%s is now disabled:\n%s", app.DisplayName, message), nil
}

func (p *Proxy) AppIsEnabled(app *apps.App) bool {
	if app.AppType == apps.AppTypeBuiltin {
		return true
	}
	if app.Disabled {
		return false
	}
	if m, _ := p.store.Manifest.Get(app.AppID); m == nil {
		return false
	}
	return true
}
