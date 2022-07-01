// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	_ "embed" // a test package, effectively
	"fmt"
	"sync/atomic"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
	"github.com/mattermost/mattermost-plugin-apps/server/builtin"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type lifecycleApp struct {
	*goapp.App
	callbacks map[string]func(*Helper, goapp.CallRequest)
	received  map[string]apps.ExpandedContext
}

var lifecycleSeq int32

func newLifecycleApp(th *Helper, install, uninstall, enable, disable bool) lifecycleApp {
	seq := atomic.AddInt32(&lifecycleSeq, 1)
	m := apps.Manifest{
		AppID:       apps.AppID(fmt.Sprintf("uninstalltest-%v", seq)),
		Version:     "v1.1.0",
		DisplayName: "This app creates data to verify that UninstallApp cleans it up",
		HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/test/restapitest",
		RequestedPermissions: apps.Permissions{
			apps.PermissionActAsUser,
			apps.PermissionActAsBot,
		},
	}
	normalExpand := apps.Expand{
		ActingUser:            apps.ExpandAll,
		ActingUserAccessToken: apps.ExpandAll,
		App:                   apps.ExpandAll,
	}
	if install {
		m.OnInstall = apps.NewCall("/install").WithExpand(normalExpand)
	}
	if uninstall {
		m.OnUninstall = apps.NewCall("/uninstall").WithExpand(normalExpand)
	}
	if enable {
		m.OnEnable = apps.NewCall("/enable").WithExpand(normalExpand)
	}
	if disable {
		m.OnDisable = apps.NewCall("/disable").WithExpand(normalExpand)
	}

	app := lifecycleApp{
		App:       goapp.MakeAppOrPanic(m),
		received:  map[string]apps.ExpandedContext{},
		callbacks: map[string]func(*Helper, goapp.CallRequest){},
	}

	handler := func(creq goapp.CallRequest) apps.CallResponse {
		app.received[creq.Path] = creq.Context.ExpandedContext

		require.NotNil(th, creq.Context.ActingUser, "must be called as the acting user")
		require.NotEmpty(th, creq.Context.ActingUser.Id)
		require.NotEmpty(th, creq.Context.ActingUserAccessToken)
		require.NotEqual(th, creq.Context.BotUserID, creq.Context.ActingUser.Id, "must be called as the user running the InstallApp API, not as the bot")

		if callback := app.callbacks[creq.Path]; callback != nil {
			callback(th, creq)
		}
		return apps.NewTextResponse(creq.Path + " called")
	}

	app.HandleCall("/install", handler)
	app.HandleCall("/uninstall", handler)
	app.HandleCall("/enable", handler)
	app.HandleCall("/disable", handler)

	return app
}

func testLifecycle(th *Helper) {
	th.Run("App with no callbacks in manifest gets none", func(th *Helper) {
		app := newLifecycleApp(th, false, false, false, false)

		th.InstallAppWithCleanup(app.App)
		require.Len(th, app.received, 0)

		th.DisableApp(app.App)
		require.Len(th, app.received, 0)

		th.EnableApp(app.App)
		require.Len(th, app.received, 0)

		th.UninstallApp(app.Manifest.AppID)
		require.Len(th, app.received, 0)
	})

	th.Run("App can get all callbacks", func(th *Helper) {
		app := newLifecycleApp(th, true, true, true, true)
		expectedContext := apps.ExpandedContext{
			// App is auto-set.
			ActingUser: th.ServerTestHelper.SystemAdminUser,
		}

		th.InstallAppWithCleanup(app.App)
		require.Len(th, app.received, 1)
		th.verifyExpandedContext(apps.ExpandAll, true, expectedContext, app.received["/install"])
		delete(app.received, "/install")

		th.DisableApp(app.App)
		require.Len(th, app.received, 1)
		th.verifyExpandedContext(apps.ExpandAll, true, expectedContext, app.received["/disable"])
		delete(app.received, "/disable")

		th.EnableApp(app.App)
		require.Len(th, app.received, 1)
		th.verifyExpandedContext(apps.ExpandAll, true, expectedContext, app.received["/enable"])
		delete(app.received, "/enable")

		th.UninstallApp(app.Manifest.AppID)
		require.Len(th, app.received, 2)
		th.verifyExpandedContext(apps.ExpandAll, true, expectedContext, app.received["/disable"])
		th.verifyExpandedContext(apps.ExpandAll, true, expectedContext, app.received["/uninstall"])
		delete(app.received, "/disable")
		delete(app.received, "/uninstall")
	})

	th.Run("Uninstall cleans up app data", func(th *Helper) {
		app := newLifecycleApp(th, true, false, false, false)
		app.callbacks["/install"] = func(th *Helper, creq goapp.CallRequest) {
			th.Run("create test data", func(th *Helper) {
				// Create KV data
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

				// Create subscriptions.
				subscribe := func(client *appclient.Client, event apps.Event) {
					err := client.Subscribe(&apps.Subscription{
						Event: event,
						Call:  *apps.NewCall("/echo"),
					})
					require.NoError(th, err)
				}
				subscribe(creq.AsBot(), apps.Event{Subject: apps.SubjectBotJoinedTeam})
				subscribe(creq.AsActingUser(), apps.Event{Subject: apps.SubjectBotJoinedTeam})
				subscribe(creq.AsActingUser(), apps.Event{Subject: apps.SubjectChannelCreated, TeamID: th.ServerTestHelper.BasicTeam.Id})
				subscribe(creq.AsActingUser(), apps.Event{Subject: apps.SubjectUserJoinedChannel, ChannelID: th.ServerTestHelper.BasicChannel.Id})
			})
		}
		th.InstallAppWithCleanup(app.App)

		infoRequest := apps.CallRequest{
			Call: *apps.NewCall(builtin.PathDebugKVInfo).WithExpand(apps.Expand{ActingUser: apps.ExpandSummary}),
		}

		cresp := th.HappyAdminCall(builtin.AppID, infoRequest)
		require.Equal(th, apps.CallResponseTypeOK, cresp.Type)
		info := store.KVDebugInfo{}
		utils.Remarshal(&info, cresp.Data)
		require.Len(th, info.Apps, 1)
		info.Total = 0
		info.ManifestCount = 0
		info.Apps[app.Manifest.AppID].AppKVCountByUserID = nil
		require.EqualValues(th, store.KVDebugInfo{
			AppsTotal:         6,
			InstalledAppCount: 1,
			OAuth2StateCount:  0,
			Other:             0, // debug clean before the test clears out the special bot key was: 1
			SubscriptionCount: 3,
			Apps: map[apps.AppID]*store.KVDebugAppInfo{
				app.Manifest.AppID: {
					AppKVCount:            4,
					AppKVCountByNamespace: map[string]int{"": 1, "p1": 2, "p2": 1},
					TokenCount:            2,
				},
			},
		}, info)

		th.UninstallApp(th.InstalledApp.AppID)
		th.Run("uninstall clears KV data", func(th *Helper) {
			cresp := th.HappyAdminCall(builtin.AppID, infoRequest)
			require.Equal(th, apps.CallResponseTypeOK, cresp.Type)
			info := store.KVDebugInfo{}
			utils.Remarshal(&info, cresp.Data)
			info.Total = 0
			info.ManifestCount = 0
			require.EqualValues(th, store.KVDebugInfo{
				Other: 0, // debug clean before the test clears out the special bot key was: 1
				Apps:  map[apps.AppID]*store.KVDebugAppInfo{},
			}, info)
		})

		// TODO: test bot account cleanup
		// TODO: test OAuth2 cleanup (server-side)
	})
}
