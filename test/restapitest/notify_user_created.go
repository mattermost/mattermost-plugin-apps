// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func testNotifyUserCreated(app *apps.App, appBotUser *model.User, received chan apps.CallRequest) func(*Helper) {
	return func(th *Helper) {
		forExpandClientCombinations(th, appBotUser, nil, nil, func(th *Helper, level apps.ExpandLevel, cl clientCombination) {
			th.subscribeAs(cl, app.AppID, apps.Event{Subject: apps.SubjectUserCreated}, expandEverything(level))
			user := th.createTestUser()

			n := <-received
			require.Empty(th, received)
			require.EqualValues(th, apps.NewCall("/notify").WithExpand(expandEverything(level)), &n.Call)

			th.verifyContext(level, app, cl.appActsAsSystemAdmin,
				apps.Context{
					Subject: apps.SubjectUserCreated,
					ExpandedContext: apps.ExpandedContext{
						App:        app,
						User:       user,
						ActingUser: cl.expectedActingUser,
						Locale:     "en",
					},
				},
				n.Context)
		})
	}
}
