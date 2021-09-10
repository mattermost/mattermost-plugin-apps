// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package uphttp

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (u *Upstream) GetStatic(app apps.App, path string) (io.ReadCloser, int, error) {
	if app.Manifest.HTTP == nil {
		return nil, http.StatusInternalServerError, errors.New("app is not available as type http")
	}
	rootURL, err := u.appRootURL(app, "/")
	if err != nil {
		return nil, http.StatusNotFound, err
	}
	url, err := utils.CleanURL(fmt.Sprintf("%s/%s/%s", rootURL, apps.StaticFolder, path))
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	client := u.httpOut.MakeClient(u.devMode)
	resp, err := client.Get(url) // nolint:bodyclose,gosec // Ignore gosec G107
	if err != nil {
		return nil, http.StatusBadGateway, errors.Wrapf(err, "failed to fetch: %s, error: %v", url, err)
	}
	return resp.Body, resp.StatusCode, nil
}
