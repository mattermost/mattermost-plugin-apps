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
	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
	"github.com/mattermost/mattermost-plugin-apps/upstream/uphttp"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upplugin"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (p *Proxy) Call(sessionID, actingUserID string, creq *apps.CallRequest) *apps.ProxyCallResponse {
	if creq.Context == nil || creq.Context.AppID == "" {
		resp := apps.NewErrorCallResponse(utils.NewInvalidError("must provide Context and set the app ID"))
		return apps.NewProxyCallResponse(resp, nil)
	}

	if actingUserID != "" {
		creq.Context.ActingUserID = actingUserID
		creq.Context.UserID = actingUserID
	}

	app, err := p.store.App.Get(creq.Context.AppID)

	var metadata *apps.AppMetadataForClient
	if app != nil {
		metadata = &apps.AppMetadataForClient{
			BotUserID:   app.BotUserID,
			BotUsername: app.BotUsername,
		}
	}

	if err != nil {
		return apps.NewProxyCallResponse(apps.NewErrorCallResponse(err), metadata)
	}

	if creq.Path[0] != '/' {
		return apps.NewProxyCallResponse(apps.NewErrorCallResponse(utils.NewInvalidError("call path must start with a %q: %q", "/", creq.Path)), metadata)
	}

	cleanPath, err := utils.CleanPath(creq.Path)
	if err != nil {
		return apps.NewProxyCallResponse(apps.NewErrorCallResponse(err), metadata)
	}
	creq.Path = cleanPath

	up, err := p.upstreamForApp(app)
	if err != nil {
		return apps.NewProxyCallResponse(apps.NewErrorCallResponse(err), metadata)
	}

	// Clear any ExpandedContext as it should always be set by an expander for security reasons
	creq.Context.ExpandedContext = apps.ExpandedContext{}

	conf := p.conf.GetConfig()
	cc := conf.SetContextDefaultsForApp(creq.Context.AppID, creq.Context)

	expander := p.newExpander(cc, p.mm, p.conf, p.store, sessionID)
	cc, err = expander.ExpandForApp(app, creq.Expand)
	if err != nil {
		return apps.NewProxyCallResponse(apps.NewErrorCallResponse(err), metadata)
	}
	clone := *creq
	clone.Context = cc

	callResponse := upstream.Call(up, &clone)

	if callResponse.Type == "" {
		callResponse.Type = apps.CallResponseTypeOK
	}

	if callResponse.Form != nil && callResponse.Form.Icon != "" {
		icon, err := normalizeStaticPath(conf, cc.AppID, callResponse.Form.Icon)
		if err != nil {
			p.mm.Log.Debug("Invalid icon path in form. Ignoring it.", "app_id", app.AppID, "icon", callResponse.Form.Icon, "error", err.Error())
			callResponse.Form.Icon = ""
		} else {
			callResponse.Form.Icon = icon
		}
	}

	return apps.NewProxyCallResponse(callResponse, metadata)
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

func (p *Proxy) Notify(cc *apps.Context, subj apps.Subject) error {
	subs, err := p.store.Subscription.Get(subj, cc.TeamID, cc.ChannelID)
	if err != nil {
		return err
	}

	expander := p.newExpander(cc, p.mm, p.conf, p.store, "")

	notify := func(sub *apps.Subscription) error {
		call := sub.Call
		if call == nil {
			return errors.New("nothing to call")
		}

		callRequest := &apps.CallRequest{Call: *call}
		app, err := p.store.App.Get(sub.AppID)
		if err != nil {
			return err
		}
		callRequest.Context, err = expander.ExpandForApp(app, callRequest.Expand)
		if err != nil {
			return err
		}
		callRequest.Context.Subject = subj

		up, err := p.upstreamForApp(app)
		if err != nil {
			return err
		}
		return upstream.Notify(up, callRequest)
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

func (p *Proxy) NotifyRemoteWebhook(app *apps.App, data []byte, webhookPath string) error {
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteWebhooks) {
		return utils.NewForbiddenError("%s does not have permission %s", app.AppID, apps.PermissionRemoteWebhooks)
	}

	up, err := p.upstreamForApp(app)
	if err != nil {
		return err
	}

	var datav interface{}
	err = json.Unmarshal(data, &datav)
	if err != nil {
		// if the data can not be decoded as JSON, send it "as is", as a string.
		datav = string(data)
	}

	// TODO: do we need to customize the Expand & State for the webhook Call?
	creq := &apps.CallRequest{
		Call: apps.Call{
			Path: path.Join(apps.PathWebhook, webhookPath),
		},
		Context: p.conf.GetConfig().SetContextDefaultsForApp(app.AppID, &apps.Context{
			ActingUserID: app.BotUserID,
		}),
		Values: map[string]interface{}{
			"data": datav,
		},
	}
	expander := p.newExpander(creq.Context, p.mm, p.conf, p.store, "")
	creq.Context, err = expander.ExpandForApp(app, creq.Expand)
	if err != nil {
		return err
	}

	return upstream.Notify(up, creq)
}

func (p *Proxy) GetAsset(appID apps.AppID, path string) (io.ReadCloser, int, error) {
	m, err := p.store.Manifest.Get(appID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, utils.ErrNotFound) {
			status = http.StatusNotFound
		}
		return nil, status, err
	}
	up, err := p.staticUpstreamForManifest(m)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return up.GetStatic(path)
}

func (p *Proxy) getAssetForApp(app *apps.App, path string) (io.ReadCloser, int, error) {
	up, err := p.staticUpstreamForApp(app)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return up.GetStatic(path)
}

func (p *Proxy) staticUpstreamForManifest(m *apps.Manifest) (upstream.StaticUpstream, error) {
	switch m.AppType {
	case apps.AppTypeHTTP:
		return uphttp.NewStaticUpstream(m, p.httpOut), nil

	case apps.AppTypeAWSLambda:
		return upaws.NewStaticUpstream(m, p.aws, p.s3AssetBucket), nil

	case apps.AppTypeBuiltin:
		return nil, errors.New("static assets are not supported for builtin apps")

	case apps.AppTypePlugin:
		app, err := p.store.App.Get(m.AppID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get app for static asset")
		}
		return upplugin.NewStaticUpstream(app, &p.mm.Plugin), nil

	default:
		return nil, utils.NewInvalidError("not a valid app type: %s", m.AppType)
	}
}

func (p *Proxy) staticUpstreamForApp(app *apps.App) (upstream.StaticUpstream, error) {
	switch app.AppType {
	case apps.AppTypeHTTP, apps.AppTypeAWSLambda, apps.AppTypeBuiltin:
		return p.staticUpstreamForManifest(&app.Manifest)

	case apps.AppTypePlugin:
		return upplugin.NewStaticUpstream(app, &p.mm.Plugin), nil

	default:
		return nil, utils.NewInvalidError("not a valid app type: %s", app.AppType)
	}
}

func (p *Proxy) upstreamForApp(app *apps.App) (upstream.Upstream, error) {
	if !p.AppIsEnabled(app) {
		return nil, errors.Errorf("%s is disabled", app.AppID)
	}
	conf := p.conf.GetConfig()
	err := isAppTypeSupported(conf, &app.Manifest)
	if err != nil {
		return nil, err
	}

	switch app.AppType {
	case apps.AppTypeHTTP:
		return uphttp.NewUpstream(app, p.httpOut), nil

	case apps.AppTypeAWSLambda:
		return upaws.NewUpstream(app, p.aws, p.s3AssetBucket), nil

	case apps.AppTypeBuiltin:
		up := p.builtinUpstreams[app.AppID]
		if up == nil {
			return nil, utils.NewNotFoundError("builtin app not found: %s", app.AppID)
		}
		return up, nil
	case apps.AppTypePlugin:
		return upplugin.NewUpstream(app, &p.mm.Plugin), nil
	default:
		return nil, utils.NewInvalidError("invalid app type: %s", app.AppType)
	}
}

func isAppTypeSupported(conf config.Config, m *apps.Manifest) error {
	supportedTypes := []apps.AppType{
		apps.AppTypeAWSLambda,
		apps.AppTypeBuiltin,
		apps.AppTypePlugin,
	}
	mode := "Mattermost Cloud"

	switch {
	case conf.DeveloperMode:
		return nil

	case conf.MattermostCloudMode:

	case !conf.MattermostCloudMode:
		// Self-managed
		supportedTypes = append(supportedTypes, apps.AppTypeHTTP)
		mode = "Self-managed"

	default:
		return errors.New("unreachable")
	}

	for _, t := range supportedTypes {
		if m.AppType == t {
			return nil
		}
	}
	return utils.NewForbiddenError("%s is not allowed in %s mode, only %s", m.AppType, mode, supportedTypes)
}
