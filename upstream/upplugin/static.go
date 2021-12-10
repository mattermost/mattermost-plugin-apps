// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upplugin

import (
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

func (u *StaticUpstream) GetStatic(app apps.App, assetPath string) (io.ReadCloser, int, error) {
	if !app.Contains(apps.DeployPlugin) {
		return nil, http.StatusInternalServerError, errors.New("app is not available as type plugin")
	}
	url := path.Join("/"+app.Manifest.Plugin.PluginID, apps.PluginAppPath, appspath.StaticFolder, assetPath)

	resp, err := u.httpClient.Get(url) // nolint:bodyclose,gosec // Ignore gosec G107
	if err != nil {
		return nil, http.StatusBadGateway, errors.Wrapf(err, "failed to fetch: %s, error: %v", url, err)
	}

	return resp.Body, resp.StatusCode, nil
}
