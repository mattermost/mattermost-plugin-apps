// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upstream

import (
	"io"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// Upstream should be abbreviated as `up`.
type Upstream interface {
	StaticUpstream
	Roundtrip(app *apps.App, call *apps.CallRequest, async bool) (io.ReadCloser, error)
}

type StaticUpstream interface {
	GetStatic(m *apps.Manifest, path string) (io.ReadCloser, int, error)
}
