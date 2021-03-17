// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"io"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type Proxy interface {
	AddBuiltinUpstream(apps.AppID, Upstream)
	AppIsEnabled(app *apps.App) bool
	Call(apps.SessionToken, *apps.CallRequest) *apps.CallResponse
	DisableApp(cc *apps.Context, app *apps.App) (md.MD, error)
	EnableApp(cc *apps.Context, app *apps.App) (md.MD, error)
	GetAsset(apps.AppID, string) (io.ReadCloser, int, error)
	GetBindings(apps.SessionToken, *apps.Context) ([]*apps.Binding, error)
	Notify(cc *apps.Context, subj apps.Subject) error
}
