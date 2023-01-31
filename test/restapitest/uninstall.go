// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	_ "embed" // a test package, effectively
	"fmt"

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
			Version:     "v1.2.0",
			DisplayName: "This app creates data to verify that UninstallApp cleans it up",
			HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/test/restapitest",
			OnInstall:   apps.NewCall("/install").ExpandActingUserClient(),
			RequestedPermissions: apps.Permissions{
				apps.PermissionActAsUser,
				apps.PermissionActAsBot,
				apps.PermissionRemoteOAuth2,
			},
		},
	)
	app.HandleCall("/install",
		func(creq goapp.CallRequest) apps.CallResponse {
			require.NotNil(th, creq.Context.ActingUser, "must be called as the acting user")
			require.NotEmpty(th, creq.Context.ActingUser.Id)
			require.NotEmpty(th, creq.Context.ActingUserAccessToken)
			require.NotEqual(th, creq.Context.BotUserID, creq.Context.ActingUser.Id, "must be called as the user running the InstallApp API, not as the bot")

			// Create KV data
			testv := map[string]any{"field": "test-value"}
			setKV := func(client *appclient.Client, prefix, key string) {
				changed, err := client.KVSet(prefix, key, testv)
				require.True(th, changed)
				require.NoError(th, err)
			}
			setKV(creq.AsBot(), "p1", "id1")
			setKV(creq.AsBot(), "", "id2")
			setKV(creq.AsActingUser(), "p1", "id1")
			setKV(creq.AsActingUser(), "p2", "id2")

			// Create subscriptions.
			subscribe := func(client *appclient.Client, event apps.Event) {
				err := client.Subscribe(&apps.Subscription{
					Event: event,
					Call:  *apps.NewCall("/echo"),
				})
				require.NoError(th, err)
			}
			subscribe(creq.AsBot(), apps.Event{Subject: apps.SubjectUserJoinedChannel})
			subscribe(creq.AsBot(), apps.Event{Subject: apps.SubjectUserJoinedTeam})
			subscribe(creq.AsActingUser(), apps.Event{Subject: apps.SubjectUserJoinedChannel})
			subscribe(creq.AsActingUser(), apps.Event{Subject: apps.SubjectChannelCreated, TeamID: th.ServerTestHelper.BasicTeam.Id})
			subscribe(creq.AsActingUser(), apps.Event{Subject: apps.SubjectUserJoinedChannel, ChannelID: th.ServerTestHelper.BasicChannel.Id})

			// Create OAuth2 data
			err := creq.AsActingUser().StoreOAuth2App(apps.OAuth2App{
				ClientID:     "testing",
				ClientSecret: "testingSecret",
			})
			require.NoError(th, err)

			setUser := func(client *appclient.Client) {
				err := client.StoreOAuth2User("testing")
				require.NoError(th, err)
			}
			setUser(creq.AsActingUser())
			return apps.NewTextResponse("installed")
		})

	return app
}

func testUninstall(th *Helper) {
	infoRequest := apps.CallRequest{
		Call: *apps.NewCall(builtin.PathDebugKVInfo).WithExpand(apps.Expand{ActingUser: apps.ExpandSummary}),
	}
	pollute := func(n int) {
		cresp := th.HappyAdminCall(builtin.AppID, apps.CallRequest{
			Call:   *apps.NewCall(builtin.PathDebugStorePollute).WithExpand(apps.Expand{ActingUser: apps.ExpandSummary}),
			Values: map[string]any{"count": fmt.Sprintf("%v", n)},
		})
		require.Equal(th, apps.CallResponseTypeOK, cresp.Type)
	}

	pollute(599)
	th.InstallAppWithCleanup(newUninstallApp(th))
	pollute(1401)

	th.Run("check test app data", func(th *Helper) {
		cresp := th.HappyAdminCall(builtin.AppID, infoRequest)
		require.Equal(th, apps.CallResponseTypeOK, cresp.Type)
		info := store.KVDebugInfo{}
		utils.Remarshal(&info, cresp.Data)
		require.Len(th, info.Apps, 1)
		require.EqualValues(th, store.KVDebugInfo{
			Apps: map[apps.AppID]*store.KVDebugAppInfo{
				th.LastInstalledApp.Manifest.AppID: {
					AppKVCount:            4,
					AppKVCountByNamespace: map[string]int{"": 1, "p1": 2, "p2": 1},
					AppKVCountByUserID:    map[string]int{th.LastInstalledApp.BotUserID: 2, th.ServerTestHelper.SystemAdminUser.Id: 2},
					TokenCount:            2,
					UserCount:             1,
				},
			},
			AppsTotal:        7,
			CachedStoreTotal: 9,
			CachedStoreCountByName: map[string]int{
				"apps":          1,
				"manifests":     1,
				"subscriptions": 4,
			},
			Debug: 2000,
			Total: 2016,
		}, info)
	})

	th.UninstallApp(uninstallID)

	th.Run("uninstall clears KV data", func(th *Helper) {
		cresp := th.HappyAdminCall(builtin.AppID, infoRequest)
		require.Equal(th, apps.CallResponseTypeOK, cresp.Type)
		info := store.KVDebugInfo{}
		utils.Remarshal(&info, cresp.Data)
		require.EqualValues(th, store.KVDebugInfo{
			Apps:             map[apps.AppID]*store.KVDebugAppInfo{},
			AppsTotal:        0,
			CachedStoreTotal: 4,
			CachedStoreCountByName: map[string]int{
				"manifests": 1,
			},
			Debug:            2000,
			OAuth2StateCount: 0,
			Other:            0,
			Total:            2004,
		}, info)
	})

	// TODO: test bot account cleanup
}
