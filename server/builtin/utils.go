// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func newUserCall(path string) *apps.Call {
	call := apps.NewCall(path)
	call.Expand = &apps.Expand{
		ActingUser:            apps.ExpandSummary,
		ActingUserAccessToken: apps.ExpandAll,
	}
	return call.WithLocale()
}
