// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package uphttp

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (u *Upstream) GetStatic(app apps.App, path string) (io.ReadCloser, int, error) {
	if app.Manifest.HTTP == nil {
		return nil, http.StatusInternalServerError, errors.New("app is not available as type http")
	}
	url := fmt.Sprintf("%s/%s/%s", app.Manifest.HTTP.RootURL, apps.StaticFolder, path)

	resp, err := http.Get(url) // nolint:bodyclose,gosec // Ignore gosec G107
	if err != nil {
		return nil, http.StatusBadGateway, errors.Wrapf(err, "failed to fetch: %s, error: %v", url, err)
	}
	return resp.Body, resp.StatusCode, nil
}
