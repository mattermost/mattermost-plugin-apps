// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package uphttp

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (u *Upstream) GetStatic(ctx context.Context, app apps.App, urlPath string) (io.ReadCloser, int, error) {
	if !app.Manifest.Contains(apps.DeployHTTP) {
		return nil, http.StatusInternalServerError, errors.New("app is not available as type http")
	}
	rootURL, err := u.appRootURL(app, "/")
	if err != nil {
		return nil, http.StatusNotFound, err
	}
	url, err := utils.CleanURL(fmt.Sprintf("%s/%s/%s", rootURL, path.StaticFolder, urlPath))
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	client := u.httpOut.MakeClient(u.devMode)
	resp, err := client.Do(req) // nolint:bodyclose,gosec // Ignore gosec G107
	if err != nil {
		return nil, http.StatusBadGateway, errors.Wrapf(err, "failed to fetch: %s, error: %v", url, err)
	}
	return resp.Body, resp.StatusCode, nil
}
