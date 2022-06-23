// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	_ "embed" // a test package, effectively

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/builtin"
)

func testMisc(th *Helper) {
	th.Run("user can not invoke builtin debug calls", func(th *Helper) {
		infoRequest := apps.CallRequest{
			Call: *apps.NewCall(builtin.PathDebugKVInfo).WithExpand(apps.Expand{
				ActingUser: apps.ExpandSummary,
			}),
			Values: map[string]interface{}{
				builtin.FieldAppID: uninstallID,
			},
		}

		cresp, _, err := th.Call(builtin.AppID, infoRequest)
		require.NoError(th, err)
		require.Equal(th, apps.CallResponseTypeError, cresp.Type)
		require.Regexp(th, `user \w+ \(\w+\) is not a sysadmin: unauthorized`, cresp.Text)
	})
}
