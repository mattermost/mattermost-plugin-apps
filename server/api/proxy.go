// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"io"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type Proxy interface {
	GetBindings(apps.SessionToken, *apps.Context) ([]*apps.Binding, error)
	Call(apps.SessionToken, *apps.CallRequest) *apps.CallResponse
	Notify(cc *apps.Context, subj apps.Subject) error

	ProvisionBuiltIn(apps.AppID, Upstream)

	GetAsset(apps.AppID, string) (io.ReadCloser, int, error)

	InvalidateCache(apps.AppID, string, string) error
	CacheDeleteAll(apps.AppID) error
	CacheDeleteAllApps() []error
}
