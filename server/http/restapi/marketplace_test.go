package restapi

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func TestMergeApps(t *testing.T) {
	for name, test := range map[string]struct {
		a        []MarketplaceApp
		b        []MarketplaceApp
		expected []MarketplaceApp
	}{
		"same slices": {
			a: []MarketplaceApp{{
				Manifest: apps.Manifest{
					AppID: "someID",
				}},
			},
			b: []MarketplaceApp{{
				Manifest: apps.Manifest{
					AppID: "someID",
				}},
			},
			expected: []MarketplaceApp{{
				Manifest: apps.Manifest{
					AppID: "someID",
				}},
			},
		},
		"a is empty": {
			a: []MarketplaceApp{},
			b: []MarketplaceApp{{
				Manifest: apps.Manifest{
					AppID: "someID",
				}},
			},
			expected: []MarketplaceApp{{
				Manifest: apps.Manifest{
					AppID: "someID",
				}},
			},
		},
		"b is empty": {
			a: []MarketplaceApp{{
				Manifest: apps.Manifest{
					AppID: "someID",
				}},
			},
			b: []MarketplaceApp{},
			expected: []MarketplaceApp{{
				Manifest: apps.Manifest{
					AppID: "someID",
				}},
			},
		},
		"both empty": {
			a:        []MarketplaceApp{},
			b:        []MarketplaceApp{},
			expected: []MarketplaceApp{},
		},
		"two different elements": {
			a: []MarketplaceApp{{
				Manifest: apps.Manifest{
					AppID: "someID",
				}},
			},
			b: []MarketplaceApp{{
				Manifest: apps.Manifest{
					AppID: "some other ID",
				}},
			},
			expected: []MarketplaceApp{
				{
					Manifest: apps.Manifest{
						AppID: "someID",
					},
				},
				{
					Manifest: apps.Manifest{
						AppID: "some other ID",
					},
				},
			},
		},
		"same element, should take first": {
			a: []MarketplaceApp{{
				Manifest: apps.Manifest{
					AppID:       "someID",
					Description: "someDescription",
				}},
			},
			b: []MarketplaceApp{{
				Manifest: apps.Manifest{
					AppID:       "someID",
					Description: "someOtherDescription",
				}},
			},
			expected: []MarketplaceApp{
				{
					Manifest: apps.Manifest{
						AppID:       "someID",
						Description: "someDescription",
					},
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			result := mergeApps(test.a, test.b)

			assert.ElementsMatch(t, test.expected, result)
		})
	}
}
