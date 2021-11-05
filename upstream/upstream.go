// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upstream

import (
	"context"
	"io"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// Upstream should be abbreviated as `up`.
type Upstream interface {
	Roundtrip(ctx context.Context, _ apps.App, _ apps.CallRequest, async bool) (io.ReadCloser, error)
	GetStatic(ctx context.Context, _ apps.App, path string) (io.ReadCloser, int, error)
}
