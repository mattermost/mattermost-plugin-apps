// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upplugin

import (
	"context"
	"io"
	"net/http"
	"path"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	appspath "github.com/mattermost/mattermost-plugin-apps/apps/path"
)

type StaticUpstream struct {
	httpClient http.Client
}

func NewStaticUpstream(api PluginHTTPAPI) *StaticUpstream {
	return &StaticUpstream{
		httpClient: MakePluginHTTPClient(api),
	}
}

func (u *StaticUpstream) GetStatic(ctx context.Context, app apps.App, assetPath string) (io.ReadCloser, int, error) {
	if !app.Contains(apps.DeployPlugin) {
		return nil, http.StatusInternalServerError, errors.New("app is not available as type plugin")
	}
	url := path.Join("/"+app.Manifest.Plugin.PluginID, apps.PluginAppPath, appspath.StaticFolder, assetPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	resp, err := u.httpClient.Do(req) // nolint:bodyclose,gosec // Ignore gosec G107
	if err != nil {
		return nil, http.StatusBadGateway, errors.Wrapf(err, "failed to fetch: %s, error: %v", url, err)
	}

	return resp.Body, resp.StatusCode, nil
}
