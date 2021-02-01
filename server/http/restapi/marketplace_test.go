package restapi

import (
	"testing"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/stretchr/testify/assert"
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
	} {
		t.Run(name, func(t *testing.T) {
			result := mergeApps(test.a, test.b)

			assert.ElementsMatch(t, test.expected, result)
		})
	}
}
