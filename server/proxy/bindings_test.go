package proxy

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

func testBinding(appID apps.AppID, parent apps.Location, n string) []apps.Binding {
	return []apps.Binding{
		{
			AppID:    appID,
			Location: parent,
			Bindings: []apps.Binding{
				{
					AppID:    appID,
					Location: apps.Location(fmt.Sprintf("id-%s", n)),
					Hint:     fmt.Sprintf("hint-%s", n),
				},
			},
		},
	}
}

func TestMergeBindings(t *testing.T) {
	type TC struct {
		name               string
		bb1, bb2, expected []apps.Binding
	}

	for _, tc := range []TC{
		{
			name: "happy simplest",
			bb1: []apps.Binding{
				{
					Location: "1",
				},
			},
			bb2: []apps.Binding{
				{
					Location: "2",
				},
			},
			expected: []apps.Binding{
				{
					Location: "1",
				},
				{
					Location: "2",
				},
			},
		},
		{
			name:     "happy simple 1",
			bb1:      testBinding("app1", apps.LocationCommand, "simple"),
			bb2:      nil,
			expected: testBinding("app1", apps.LocationCommand, "simple"),
		},
		{
			name:     "happy simple 2",
			bb1:      nil,
			bb2:      testBinding("app1", apps.LocationCommand, "simple"),
			expected: testBinding("app1", apps.LocationCommand, "simple"),
		},
		{
			name:     "happy simple same",
			bb1:      testBinding("app1", apps.LocationCommand, "simple"),
			bb2:      testBinding("app1", apps.LocationCommand, "simple"),
			expected: testBinding("app1", apps.LocationCommand, "simple"),
		},
		{
			name: "happy simple merge",
			bb1:  testBinding("app1", apps.LocationPostMenu, "simple"),
			bb2:  testBinding("app1", apps.LocationCommand, "simple"),
			expected: append(
				testBinding("app1", apps.LocationPostMenu, "simple"),
				testBinding("app1", apps.LocationCommand, "simple")...,
			),
		},
		{
			name: "happy simple 2 apps",
			bb1:  testBinding("app1", apps.LocationCommand, "simple"),
			bb2:  testBinding("app2", apps.LocationCommand, "simple"),
			expected: append(
				testBinding("app1", apps.LocationCommand, "simple"),
				testBinding("app2", apps.LocationCommand, "simple")...,
			),
		},
		{
			name: "happy 2 simple commands",
			bb1:  testBinding("app1", apps.LocationCommand, "simple1"),
			bb2:  testBinding("app1", apps.LocationCommand, "simple2"),
			expected: []apps.Binding{
				{
					AppID:    "app1",
					Location: "/command",
					Bindings: []apps.Binding{
						{
							AppID:    "app1",
							Location: "id-simple1",
							Hint:     "hint-simple1",
						},
						{
							AppID:    "app1",
							Location: "id-simple2",
							Hint:     "hint-simple2",
						},
					},
				},
			},
		},
		{
			name: "happy 2 apps",
			bb1: []apps.Binding{
				{
					Location: "/post_menu",
					Bindings: []apps.Binding{
						{
							AppID:       "zendesk",
							Label:       "Create zendesk ticket",
							Description: "Create ticket in zendesk",
							Form: &apps.Form{
								Submit: apps.NewCall("http://localhost:4000/create"),
							},
						},
					},
				},
			},
			bb2: []apps.Binding{
				{
					Location: "/post_menu",
					Bindings: []apps.Binding{
						{
							AppID:       "hello",
							Label:       "Create hello ticket",
							Description: "Create ticket in hello",
							Form: &apps.Form{
								Submit: apps.NewCall("http://localhost:4000/hello"),
							},
						},
					},
				},
			},
			expected: []apps.Binding{
				{
					Location: "/post_menu",
					Bindings: []apps.Binding{
						{
							AppID:       "zendesk",
							Label:       "Create zendesk ticket",
							Description: "Create ticket in zendesk",
							Form: &apps.Form{
								Submit: apps.NewCall("http://localhost:4000/create"),
							},
						},
						{
							AppID:       "hello",
							Label:       "Create hello ticket",
							Description: "Create ticket in hello",
							Form: &apps.Form{
								Submit: apps.NewCall("http://localhost:4000/hello"),
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := mergeBindings(tc.bb1, tc.bb2)
			equalBindings(t, tc.expected, out)
		})
	}
}

// equalBindings asserts that two slices of bindings are equal ignoring the order of the elements.
// If there are duplicate elements, the number of appearances of each of them in both lists should match.
//
// equalBindings calls t.Fail if the elements not match.
func equalBindings(t *testing.T, expected, actual []apps.Binding) {
	opt := cmpopts.SortSlices(func(a apps.Binding, b apps.Binding) bool {
		return a.AppID < b.AppID
	})

	if diff := cmp.Diff(expected, actual, opt); diff != "" {
		t.Errorf("Bindings mismatch (-expected +actual):\n%s", diff)
	}
}

func TestCleanAppBinding(t *testing.T) {
	app := &apps.App{
		Manifest: apps.Manifest{
			AppID: "appid",
		},
		GrantedLocations: apps.Locations{
			apps.LocationCommand,
			apps.LocationChannelHeader,
		},
	}

	type TC struct {
		in               apps.Binding
		locPrefix        apps.Location
		userAgent        string
		expected         *apps.Binding
		expectedProblems string
	}

	for name, tc := range map[string]TC{
		"happy simplest": {
			in: apps.Binding{
				Location: "test",
				Submit:   apps.NewCall("/hello"),
			},
			locPrefix: apps.LocationCommand.Sub("main-command"),
			expected: &apps.Binding{
				AppID:    "appid",
				Location: "test",
				Label:    "test",
				Submit:   apps.NewCall("/hello"),
			},
		},
		"trim location": {
			in: apps.Binding{
				Location: " test-1 \t",
				Submit:   apps.NewCall("/hello"),
			},
			locPrefix: apps.LocationCommand.Sub("main-command"),
			expected: &apps.Binding{
				AppID:    "appid",
				Location: "test-1",
				Label:    "test-1",
				Submit:   apps.NewCall("/hello"),
			},
			expectedProblems: "1 error occurred:\n\t* /command/main-command/test-1: trimmed whitespace from location\n\n",
		},
		"ERROR location PostMenu not granted": {
			in: apps.Binding{
				Location: "test",
				Submit:   apps.NewCall("/hello"),
			},
			locPrefix:        apps.LocationPostMenu,
			expected:         nil,
			expectedProblems: "1 error occurred:\n\t* /post_menu/test: location is not granted: forbidden\n\n",
		},
		"trim command label": {
			in: apps.Binding{
				Location: "test",
				Label:    "\ntest-label \t",
				Submit:   apps.NewCall("/hello"),
			},
			locPrefix: apps.LocationCommand.Sub("main-command"),
			expected: &apps.Binding{
				AppID:    "appid",
				Location: "test",
				Label:    "test-label",
				Submit:   apps.NewCall("/hello"),
			},
			expectedProblems: "1 error occurred:\n\t* /command/main-command/test: trimmed whitespace from label test-label\n\n",
		},
		"label defaults to location for command": {
			in: apps.Binding{
				Location: "test",
				Submit:   apps.NewCall("/hello"),
			},
			locPrefix: apps.LocationCommand.Sub("main-command"),
			expected: &apps.Binding{
				AppID:    "appid",
				Location: "test",
				Label:    "test",
				Submit:   apps.NewCall("/hello"),
			},
		},
		"label does not default for non-commands": {
			in: apps.Binding{
				Location: "test",
				Submit:   apps.NewCall("/hello"),
			},
			locPrefix: apps.LocationChannelHeader.Sub("some"),
			expected: &apps.Binding{
				AppID:    "appid",
				Location: "test",
				Submit:   apps.NewCall("/hello"),
			},
		},
		"ERROR neither location nor label": {
			in: apps.Binding{
				Submit: apps.NewCall("/hello"),
			},
			locPrefix:        apps.LocationCommand.Sub("main-command"),
			expected:         nil,
			expectedProblems: "1 error occurred:\n\t* /command/main-command: sub-binding with no location nor label\n\n",
		},
		"ERROR whitsepace in command label": {
			in: apps.Binding{
				Location: "test",
				Label:    "test label",
				Submit:   apps.NewCall("/hello"),
			},
			locPrefix:        apps.LocationCommand.Sub("main-command"),
			expected:         nil,
			expectedProblems: "1 error occurred:\n\t* /command/main-command/test: command label \"test label\" has multiple words\n\n",
		},
		"normalize icon path": {
			in: apps.Binding{
				Location: "test",
				Submit:   apps.NewCall("/hello"),
				Icon:     "a///static.icon",
			},
			locPrefix: apps.LocationCommand.Sub("main-command"),
			expected: &apps.Binding{
				AppID:    "appid",
				Location: "test",
				Label:    "test",
				Icon:     "/apps/appid/static/a/static.icon",
				Submit:   apps.NewCall("/hello"),
			},
		},
		"invalid icon path": {
			in: apps.Binding{
				Submit:   apps.NewCall("/hello"),
				Location: "test",
				Icon:     "../a/...//static.icon",
			},
			locPrefix: apps.LocationCommand.Sub("main-command"),
			expected: &apps.Binding{
				AppID:    "appid",
				Location: "test",
				Label:    "test",
				Submit:   apps.NewCall("/hello"),
			},
			expectedProblems: "1 error occurred:\n\t* /command/main-command/test: invalid icon path \"../a/...//static.icon\" in binding\n\n",
		},
		"ERROR: icon required for ChannelHeader in webapp": {
			in: apps.Binding{
				Location: "test",
				Submit:   apps.NewCall("/hello"),
			},
			locPrefix:        apps.LocationChannelHeader,
			userAgent:        "webapp",
			expected:         nil,
			expectedProblems: "1 error occurred:\n\t* /channel_header/test: no icon in channel header binding\n\n",
		},
		"icon not required for ChannelHeader in mobile": {
			in: apps.Binding{
				Location: "test",
				Submit:   apps.NewCall("/hello"),
			},
			locPrefix: apps.LocationChannelHeader,
			userAgent: "something-else",
			expected: &apps.Binding{
				AppID:    "appid",
				Location: "test",
				Submit:   apps.NewCall("/hello"),
			},
		},
		"ERROR: no submit/form/bindings": {
			in: apps.Binding{
				Location: "test",
			},
			locPrefix:        apps.LocationChannelHeader,
			expected:         nil,
			expectedProblems: "1 error occurred:\n\t* /channel_header/test: (only) one of  \"submit\", \"form\", or \"bindings\" must be set in a binding\n\n",
		},
		"ERROR: submit and form": {
			in: apps.Binding{
				Location: "test",
				Submit:   apps.NewCall("/hello"),
				Form:     apps.NewBlankForm(apps.NewCall("/hello")),
			},
			locPrefix:        apps.LocationChannelHeader,
			expected:         nil,
			expectedProblems: "1 error occurred:\n\t* /channel_header/test: (only) one of  \"submit\", \"form\", or \"bindings\" must be set in a binding\n\n",
		},
		"ERROR: submit and bindings": {
			in: apps.Binding{
				Location: "test",
				Submit:   apps.NewCall("/hello"),
				Bindings: []apps.Binding{
					{
						Location: "test1",
					},
					{
						Location: "test2",
					},
				},
			},
			locPrefix:        apps.LocationChannelHeader,
			expected:         nil,
			expectedProblems: "1 error occurred:\n\t* /channel_header/test: (only) one of  \"submit\", \"form\", or \"bindings\" must be set in a binding\n\n",
		},
		"clean sub-bindings": {
			in: apps.Binding{
				Location: "test",
				Bindings: []apps.Binding{
					{
						Location: "test1",
						Submit:   apps.NewCall("/hello"),
					},
					{
						Location: "test2",
						Submit:   apps.NewCall("/hello"),
					},
				},
			},
			locPrefix: apps.LocationChannelHeader,
			expected: &apps.Binding{
				AppID:    "appid",
				Location: "test",
				Bindings: []apps.Binding{
					{
						AppID:    "appid",
						Location: "test1",
						Submit:   apps.NewCall("/hello"),
					},
					{
						AppID:    "appid",
						Location: "test2",
						Submit:   apps.NewCall("/hello"),
					},
				},
			},
		},
		"clean form": {
			in: apps.Binding{
				Location: "test",
				Form: &apps.Form{
					Submit: apps.NewCall("/hello"),
					Fields: []apps.Field{
						{Name: "in valid"},
					},
				},
			},
			locPrefix: apps.LocationChannelHeader,
			expected: &apps.Binding{
				AppID:    "appid",
				Location: "test",
				Form: &apps.Form{
					Submit: apps.NewCall("/hello"),
					Fields: []apps.Field{},
				},
			},
			expectedProblems: "1 error occurred:\n\t* field name must be a single word: \"in valid\"\n\n",
		},
	} {
		t.Run(name, func(t *testing.T) {
			b, err := cleanAppBinding(app, tc.in, tc.locPrefix, tc.userAgent, config.Config{})
			if tc.expectedProblems != "" {
				require.Error(t, err)
				require.Equal(t, tc.expectedProblems, err.Error())
			} else {
				require.NoError(t, err)
				require.EqualValues(t, tc.expected, b)
			}
		})
	}
}
