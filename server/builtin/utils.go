// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func newUserCall(path string) *apps.Call {
	return apps.NewCall(path).WithExpand(apps.Expand{
		ActingUser: apps.ExpandSummary.Required(),
		Locale:     apps.ExpandAll,
	})
}
