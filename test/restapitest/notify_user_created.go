// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// Creates a test user in the Basic team. All subscribers are members of the
// team, so get can subscribe and get notified, expanding User at the requested
// level.
func notifyUserCreated(*Helper) *notifyTestCase {
	return &notifyTestCase{
		event: func(*Helper, apps.ExpandedContext) apps.Event {
			return apps.Event{
				Subject: apps.SubjectUserCreated,
			}
		},
		trigger: func(th *Helper, _ apps.ExpandedContext) apps.ExpandedContext {
			return apps.ExpandedContext{
				User: th.createTestUser(),
			}
		},
		expected: func(th *Helper, _ apps.ExpandLevel, _ appClient, data apps.ExpandedContext) apps.ExpandedContext {
			return apps.ExpandedContext{
				User: th.getUser(data.User.Id),
			}
		},
	}
}
