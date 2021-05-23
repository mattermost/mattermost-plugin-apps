// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upstream

import (
	"io"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// Upstream should be abbreviated as `up`.
type Upstream interface {
	Roundtrip(call *apps.CallRequest, async bool) (io.ReadCloser, error)
	GetStatic(path string) (io.ReadCloser, int, error)
}
