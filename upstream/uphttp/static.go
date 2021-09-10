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
	url := fmt.Sprintf("%s/%s/%s", app.Manifest.HTTPRootURL, apps.StaticFolder, path)

	client := u.httpOut.MakeClient(u.devMode)

	resp, err := client.Get(url) // nolint:bodyclose,gosec // Ignore gosec G107
	if err != nil {
		return nil, http.StatusBadGateway, errors.Wrapf(err, "failed to fetch: %s, error: %v", url, err)
	}
	return resp.Body, resp.StatusCode, nil
}
