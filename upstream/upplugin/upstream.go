// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upplugin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type Upstream struct {
	StaticUpstream
}

var _ upstream.Upstream = (*Upstream)(nil)

func NewUpstream(api PluginHTTPAPI) *Upstream {
	staticUp := NewStaticUpstream(api)
	return &Upstream{
		StaticUpstream: *staticUp,
	}
}

func (u *Upstream) Roundtrip(ctx context.Context, app apps.App, creq apps.CallRequest, async bool) (io.ReadCloser, error) {
	if async {
		go func() {
			resp, _ := u.invoke(context.Background(), app, creq)
			if resp != nil {
				resp.Body.Close()
			}
		}()
		return nil, nil
	}

	resp, err := u.invoke(ctx, app, creq) // nolint:bodyclose
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (u *Upstream) invoke(ctx context.Context, app apps.App, creq apps.CallRequest) (*http.Response, error) {
	if !app.Contains(apps.DeployPlugin) {
		return nil, errors.New("app is not available as type plugin")
	}

	return u.post(ctx, path.Join("/"+app.Manifest.Plugin.PluginID, apps.PluginAppPath, creq.Path), creq)
}

// post does not close resp.Body, it's the caller's responsibility
func (u *Upstream) post(ctx context.Context, url string, msg interface{}) (*http.Response, error) {
	piper, pipew := io.Pipe()
	go func() {
		encodeErr := json.NewEncoder(pipew).Encode(msg)
		if encodeErr != nil {
			_ = pipew.CloseWithError(encodeErr)
		}
		pipew.Close()
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, piper)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := u.httpClient.Do(req)
	switch {
	case err != nil:
		return nil, err

	case resp.StatusCode == http.StatusNotFound:
		return nil, utils.NewNotFoundError(err)

	case resp.StatusCode != http.StatusOK:
		bb, _ := httputils.ReadAndClose(resp.Body)
		return nil, errors.New(string(bb))
	}

	return resp, nil
}
