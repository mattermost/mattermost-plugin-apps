package apps

import (
	"fmt"
	"testing"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/stretchr/testify/require"
)

func testBinding(appID api.AppID, parent api.LocationID, n string) []*api.Binding {
	return []*api.Binding{
		{
			AppID:      appID,
			LocationID: parent,
			Bindings: []*api.Binding{
				{
					AppID:      appID,
					LocationID: api.LocationID(fmt.Sprintf("id-%s", n)),
					Hint:       fmt.Sprintf("hint-%s", n),
				},
			},
		},
	}
}

func TestMergeBindings(t *testing.T) {
	type TC struct {
		name               string
		bb1, bb2, expected []*api.Binding
	}

	for _, tc := range []TC{
		{
			name: "happy simplest",
			bb1: []*api.Binding{
				{
					LocationID: "1",
				},
			},
			bb2: []*api.Binding{
				{
					LocationID: "2",
				},
			},
			expected: []*api.Binding{
				{
					LocationID: "1",
				},
				{
					LocationID: "2",
				},
			},
		},
		{
			name:     "happy simple 1",
			bb1:      testBinding("app1", api.LocationCommand, "simple"),
			bb2:      nil,
			expected: testBinding("app1", api.LocationCommand, "simple"),
		},
		{
			name:     "happy simple 2",
			bb1:      nil,
			bb2:      testBinding("app1", api.LocationCommand, "simple"),
			expected: testBinding("app1", api.LocationCommand, "simple"),
		},
		{
			name:     "happy simple same",
			bb1:      testBinding("app1", api.LocationCommand, "simple"),
			bb2:      testBinding("app1", api.LocationCommand, "simple"),
			expected: testBinding("app1", api.LocationCommand, "simple"),
		},
		{
			name: "happy simple merge",
			bb1:  testBinding("app1", api.LocationPostMenu, "simple"),
			bb2:  testBinding("app1", api.LocationCommand, "simple"),
			expected: append(
				testBinding("app1", api.LocationPostMenu, "simple"),
				testBinding("app1", api.LocationCommand, "simple")...,
			),
		},
		{
			name: "happy simple 2 apps",
			bb1:  testBinding("app1", api.LocationCommand, "simple"),
			bb2:  testBinding("app2", api.LocationCommand, "simple"),
			expected: append(
				testBinding("app1", api.LocationCommand, "simple"),
				testBinding("app2", api.LocationCommand, "simple")...,
			),
		},
		{
			name: "happy 2 simple commands",
			bb1:  testBinding("app1", api.LocationCommand, "simple1"),
			bb2:  testBinding("app1", api.LocationCommand, "simple2"),
			expected: []*api.Binding{
				{
					AppID:      "app1",
					LocationID: "/command",
					Bindings: []*api.Binding{
						{
							AppID:      "app1",
							LocationID: "id-simple1",
							Hint:       "hint-simple1",
						},
						{
							AppID:      "app1",
							LocationID: "id-simple2",
							Hint:       "hint-simple2",
						},
					},
				},
			},
		},
		{
			name: "happy 2 apps",
			bb1: []*api.Binding{
				{
					LocationID: "/post_menu",
					Bindings: []*api.Binding{
						{
							AppID:       "zendesk",
							Label:       "Create zendesk ticket",
							Description: "Create ticket in zendesk",
							Call: &api.Call{
								URL: "http://localhost:4000/create",
							},
						},
					},
				},
			},
			bb2: []*api.Binding{
				{
					LocationID: "/post_menu",
					Bindings: []*api.Binding{
						{
							AppID:       "hello",
							Label:       "Create hello ticket",
							Description: "Create ticket in hello",
							Call: &api.Call{
								URL: "http://localhost:4000/hello",
							},
						},
					},
				},
			},
			expected: []*api.Binding{
				{
					LocationID: "/post_menu",
					Bindings: []*api.Binding{
						{
							AppID:       "zendesk",
							Label:       "Create zendesk ticket",
							Description: "Create ticket in zendesk",
							Call: &api.Call{
								URL: "http://localhost:4000/create",
							},
						},
						{
							AppID:       "hello",
							Label:       "Create hello ticket",
							Description: "Create ticket in hello",
							Call: &api.Call{
								URL: "http://localhost:4000/hello",
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := mergeBindings(tc.bb1, tc.bb2)
			require.Equal(t, tc.expected, out)
		})
	}
}

// []*api.Binding{
// 	{
// 		LocationID: api.LocationCommand,
// 		Bindings: []*api.Binding{
// 			{
// 				LocationID:  "message",
// 				Hint:        "[--user] message",
// 				Description: "send a message to a user",
// 				Call:        h.makeCall(PathMessage),
// 			}, {
// 				LocationID:  "manage",
// 				Hint:        "subscribe | unsubscribe ",
// 				Description: "manage channel subscriptions to greet new users",
// 				Bindings: []*api.Binding{
// 					{
// 						LocationID:  "subscribe",
// 						Hint:        "[--channel]",
// 						Description: "subscribes a channel to greet new users",
// 						Call:        h.makeCall(PathMessage, "mode", "on"),
// 					}, {
// 						LocationID:  "unsubscribe",
// 						Hint:        "[--channel]",
// 						Description: "unsubscribes a channel from greeting new users",
// 						Call:        h.makeCall(PathMessage, "mode", "off"),
// 					},
// 				},
// 			},
// 		},
// 	}, {
// 		LocationID: api.LocationPostMenu,
// 		Bindings: []*api.Binding{
// 			{
// 				LocationID:  "message",
// 				Description: "message a user",
// 				Call:        h.makeCall(PathMessage),
// 			},
// 		},
// 	},
// })
