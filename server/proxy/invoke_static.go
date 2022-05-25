// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

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

func (p *Proxy) InvokeGetStatic(r *incoming.Request, path string) (io.ReadCloser, int, error) {
	app, err := p.getEnabledDestination(r)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, utils.ErrNotFound) {
			status = http.StatusNotFound
		}
		r.Log = r.Log.WithError(err)
		return nil, status, err
	}

	return p.getStatic(r, app, path)
}

func (p *Proxy) getStatic(r *incoming.Request, app *apps.App, path string) (io.ReadCloser, int, error) {
	up, err := p.upstreamForApp(app)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return up.GetStatic(r.Ctx(), *app, path)
}

// pingApp checks if the app is accessible. Call its ping path with nothing
// expanded, ignore 404 errors coming back and consider everything else a
// "success".
func (p *Proxy) pingApp(ctx context.Context, app *apps.App) (reachable bool) {
	ctx, cancel := context.WithTimeout(ctx, pingAppTimeout)
	defer cancel()

	up, err := p.upstreamForApp(app)
	if err != nil {
		return false
	}

	_, err = upstream.Call(ctx, up, *app, apps.CallRequest{
		Call: apps.DefaultPing,
	})
	return err == nil || errors.Cause(err) == utils.ErrNotFound
}
