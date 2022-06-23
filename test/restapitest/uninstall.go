// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	_ "embed" // a test package, effectively

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
	"github.com/mattermost/mattermost-plugin-apps/server/builtin"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const uninstallID = apps.AppID("uninstalltest")

func newUninstallApp(th *Helper) *goapp.App {
	app := goapp.MakeAppOrPanic(
		apps.Manifest{
			AppID:       uninstallID,
			Version:     "v1.1.0",
			DisplayName: "This app creates data to verify that UninstallApp cleans it up",
			HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/test/restapitest",
			OnInstall:   apps.NewCall("/install").ExpandActingUserClient(),
			RequestedPermissions: apps.Permissions{
				apps.PermissionActAsUser,
				apps.PermissionActAsBot,
			},
		},
	)
	app.HandleCall("/install",
		func(creq goapp.CallRequest) apps.CallResponse {
			require.NotNil(th, creq.Context.ActingUser, "must be called as the acting user")
			require.NotEmpty(th, creq.Context.ActingUser.Id)
			require.NotEmpty(th, creq.Context.ActingUserAccessToken)
			require.NotEqual(th, creq.Context.BotUserID, creq.Context.ActingUser.Id, "must be called as the user running the InstallApp API, not as the bot")

			// create KV data
			testv := map[string]interface{}{"field": "test-value"}
			setKV := func(client *appclient.Client, prefix, key string) {
				changed, err := client.KVSet(prefix, key, testv)
				require.True(th, changed)
				require.NoError(th, err)
			}
			setKV(creq.AsBot(), "p1", "id1")
			setKV(creq.AsBot(), "", "id2")
			setKV(creq.AsActingUser(), "p1", "id1")
			setKV(creq.AsActingUser(), "p2", "id2")

			return apps.NewTextResponse("installed")
		})

	return app
}

func testUninstall(th *Helper) {
	infoRequest := apps.CallRequest{
		Call: *apps.NewCall(builtin.PathDebugKVInfo).WithExpand(apps.Expand{
			ActingUser: apps.ExpandSummary,
		}),
		Values: map[string]interface{}{
			builtin.FieldAppID: uninstallID,
		},
	}

	th.InstallAppWithCleanup(newUninstallApp(th))
	cresp := th.HappyAdminCall(builtin.AppID, infoRequest)
	require.Equal(th, apps.CallResponseTypeOK, cresp.Type)
	appInfo := store.KVDebugAppInfo{}
	utils.Remarshal(&appInfo, cresp.Data)
	require.Equal(th, 4, appInfo.AppCount)
	require.Equal(th, map[string]int{"": 1, "p1": 2, "p2": 1}, appInfo.AppByNamespace)
	require.Len(th, appInfo.AppByUserID, 2)
	require.Equal(th, 2, appInfo.TokenCount)

	th.UninstallApp(uninstallID)
	cresp, _, err := th.AdminCall(builtin.AppID, infoRequest)
	require.NoError(th, err)
	require.Equal(th, apps.CallResponseTypeError, cresp.Type)
	require.Equal(th, "uninstalltest: not found", cresp.Text)
}
