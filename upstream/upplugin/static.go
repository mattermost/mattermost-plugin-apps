// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upplugin

import (
	"io"
	"net/http"
	"path"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
)

type StaticUpstream struct {
	pluginID   string
	httpClient http.Client
}

var _ upstream.StaticUpstream = (*StaticUpstream)(nil)

func NewStaticUpstream(a *apps.App, api PluginHTTPAPI) *StaticUpstream {
	return &StaticUpstream{
		pluginID:   a.PluginID,
		httpClient: MakePluginHTTPClient(api),
	}
}

func (u *StaticUpstream) GetStatic(p string) (io.ReadCloser, int, error) {
	url := path.Join("/"+u.pluginID, apps.PluginAppPath, apps.StaticFolder, p)

	resp, err := u.httpClient.Get(url) // nolint:bodyclose,gosec
	if err != nil {
		return nil, http.StatusBadGateway, errors.Wrapf(err, "failed to fetch: %s, error: %v", url, err)
	}

	return resp.Body, resp.StatusCode, nil
}
