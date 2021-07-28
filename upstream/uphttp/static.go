// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package uphttp

import (
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
)

type StaticUpstream struct {
	httpOut httpout.Service
}

var _ upstream.StaticUpstream = (*StaticUpstream)(nil)

func NewStaticUpstream(httpOut httpout.Service) *StaticUpstream {
	return &StaticUpstream{
		httpOut: httpOut,
	}
}

func (u *StaticUpstream) GetStatic(m *apps.Manifest, path string) (io.ReadCloser, int, error) {
	if m.HTTP == nil {
		return nil, http.StatusInternalServerError, errors.New("App is not available as type http")
	}
	url := fmt.Sprintf("%s/%s/%s", m.HTTP.RootURL, apps.StaticFolder, path)

	resp, err := http.Get(url) // nolint:bodyclose,gosec // Ignore gosec G107
	if err != nil {
		return nil, http.StatusBadGateway, errors.Wrapf(err, "failed to fetch: %s, error: %v", url, err)
	}
	return resp.Body, resp.StatusCode, nil
}
