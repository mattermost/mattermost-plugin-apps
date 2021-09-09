// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upplugin

import (
	"io"
	"net/http"
	"path"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type StaticUpstream struct {
	httpClient http.Client
}

func NewStaticUpstream(api PluginHTTPAPI) *StaticUpstream {
	return &StaticUpstream{
		httpClient: MakePluginHTTPClient(api),
	}
}

func (u *StaticUpstream) GetStatic(m apps.Manifest, assetPath string) (io.ReadCloser, int, error) {
	url := path.Join("/"+m.PluginID, apps.PluginAppPath, apps.StaticFolder, assetPath)

	resp, err := u.httpClient.Get(url) // nolint:bodyclose,gosec // Ignore gosec G107
	if err != nil {
		return nil, http.StatusBadGateway, errors.Wrapf(err, "failed to fetch: %s, error: %v", url, err)
	}

	return resp.Body, resp.StatusCode, nil
}
