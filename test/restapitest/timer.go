// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/api4"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
)

const timerID = apps.AppID("timer_test")

func newTimerApp(_ testing.TB) *goapp.App {
	app := goapp.MakeAppOrPanic(
		apps.Manifest{
			AppID:       timerID,
			Version:     "v1.1.0",
			DisplayName: "tests app timers",
			HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/test/restapitest",
			RequestedPermissions: []apps.Permission{
				apps.PermissionActAsBot,
				apps.PermissionActAsUser,
			},
		},
	)

	return app
}

func testTimer(th *Helper) {
	app := newTimerApp(th.T)
	th.InstallAppWithCleanup(app)

	baseTimer := apps.Timer{
		Call: apps.Call{
			Path: "/timer/execute",
		},
	}

	th.Run("Unauthenticated requests are rejected", func(th *Helper) {
		assert := assert.New(th)
		client := th.CreateUnauthenticatedClientPP()

		resp, err := client.CreateTimer(&baseTimer)
		assert.Error(err)
		api4.CheckUnauthorizedStatus(th, resp)
	})

	th.Run("Invalid requests are rejected", func(th *Helper) {
		assert := assert.New(th)
		client := th.UserClientApp

		// No call defined
		err := client.CreateTimer(&apps.Timer{})
		assert.Error(err)

		t := baseTimer

		// Negative time
		t.At = -100
		err = client.CreateTimer(&t)
		assert.Error(err)

		// at is now
		t.At = time.Now().UnixMilli()
		err = client.CreateTimer(&t)
		assert.Error(err)

		// at is less then a second in the future
		t.At = time.Now().Add(500 * time.Millisecond).UnixMilli()
		err = client.CreateTimer(&t)
		assert.Error(err)
	})

	th.Run("Assert that timer is called", func(th *Helper) {
		assert := assert.New(th)
		client := th.UserClientApp

		var mut sync.Mutex
		var called bool
		app.HandleCall("/timer/execute",
			func(creq goapp.CallRequest) apps.CallResponse {
				mut.Lock()
				defer mut.Unlock()
				called = true
				return apps.NewTextResponse("OK")
			})

		t := baseTimer

		t.At = time.Now().Add(2 * time.Second).UnixMilli()
		err := client.CreateTimer(&t)
		assert.NoError(err)

		// Check if callback was called
		require.Eventually(th, func() bool { mut.Lock(); defer mut.Unlock(); return called }, 10*time.Second, 50*time.Millisecond)
	})
}
