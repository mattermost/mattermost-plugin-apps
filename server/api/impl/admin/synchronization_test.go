package admin

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func TestMergeApps(t *testing.T) {
	app1v1 := &apps.App{AppID: "app1", Manifest: &apps.Manifest{AppID: "app1", Version: "1"}}
	app2v1 := &apps.App{AppID: "app2", Manifest: &apps.Manifest{AppID: "app2", Version: "1"}}
	app2v2 := &apps.App{AppID: "app2", Manifest: &apps.Manifest{AppID: "app2", Version: "2"}}

	for _, tc := range []struct {
		name                   string
		appsForRegistration    apps.AppVersionMap
		registeredApps         []*apps.App
		expectedAppsToRegister apps.AppVersionMap
		expectedAppsToUpgrade  map[apps.AppID]appOldVersion
		expectedAppsToRemove   map[apps.AppID]*apps.App
	}{
		{
			name:                   "empty",
			appsForRegistration:    apps.AppVersionMap{},
			registeredApps:         []*apps.App{},
			expectedAppsToRegister: apps.AppVersionMap{},
			expectedAppsToUpgrade:  map[apps.AppID]appOldVersion{},
			expectedAppsToRemove:   map[apps.AppID]*apps.App{},
		},
		{
			name:                   "no new apps",
			appsForRegistration:    apps.AppVersionMap{"app1": "1", "app2": "1"},
			registeredApps:         []*apps.App{app1v1, app2v1},
			expectedAppsToRegister: apps.AppVersionMap{},
			expectedAppsToUpgrade:  map[apps.AppID]appOldVersion{},
			expectedAppsToRemove:   map[apps.AppID]*apps.App{},
		},
		{
			name:                   "app to register",
			appsForRegistration:    apps.AppVersionMap{"app1": "1", "app2": "1"},
			registeredApps:         []*apps.App{app1v1},
			expectedAppsToRegister: apps.AppVersionMap{"app2": "1"},
			expectedAppsToUpgrade:  map[apps.AppID]appOldVersion{},
			expectedAppsToRemove:   map[apps.AppID]*apps.App{},
		},
		{
			name:                   "apps to register",
			appsForRegistration:    apps.AppVersionMap{"app1": "1", "app2": "1"},
			registeredApps:         []*apps.App{},
			expectedAppsToRegister: apps.AppVersionMap{"app1": "1", "app2": "1"},
			expectedAppsToUpgrade:  map[apps.AppID]appOldVersion{},
			expectedAppsToRemove:   map[apps.AppID]*apps.App{},
		},
		{
			name:                   "app to upgrade",
			appsForRegistration:    apps.AppVersionMap{"app1": "2", "app2": "1"},
			registeredApps:         []*apps.App{app1v1, app2v1},
			expectedAppsToRegister: apps.AppVersionMap{},
			expectedAppsToUpgrade:  map[apps.AppID]appOldVersion{"app1": {oldApp: app1v1, newVersion: "2"}},
			expectedAppsToRemove:   map[apps.AppID]*apps.App{},
		},
		{
			name:                   "upgrade and downgrade",
			appsForRegistration:    apps.AppVersionMap{"app1": "2", "app2": "1"},
			registeredApps:         []*apps.App{app1v1, app2v2},
			expectedAppsToRegister: apps.AppVersionMap{},
			expectedAppsToUpgrade: map[apps.AppID]appOldVersion{
				"app1": {oldApp: app1v1, newVersion: "2"},
				"app2": {oldApp: app2v2, newVersion: "1"},
			},
			expectedAppsToRemove: map[apps.AppID]*apps.App{},
		},
		{
			name:                   "app to remove",
			appsForRegistration:    apps.AppVersionMap{"app1": "1"},
			registeredApps:         []*apps.App{app1v1, app2v1},
			expectedAppsToRegister: apps.AppVersionMap{},
			expectedAppsToUpgrade:  map[apps.AppID]appOldVersion{},
			expectedAppsToRemove:   map[apps.AppID]*apps.App{"app2": app2v1},
		},
		{
			name:                   "apps to remove",
			appsForRegistration:    apps.AppVersionMap{},
			registeredApps:         []*apps.App{app1v1, app2v1},
			expectedAppsToRegister: apps.AppVersionMap{},
			expectedAppsToUpgrade:  map[apps.AppID]appOldVersion{},
			expectedAppsToRemove:   map[apps.AppID]*apps.App{"app2": app2v1, "app1": app1v1},
		},
		{
			name:                   "mixed",
			appsForRegistration:    apps.AppVersionMap{"app2": "1", "app1": "2"},
			registeredApps:         []*apps.App{app1v1},
			expectedAppsToRegister: apps.AppVersionMap{"app2": "1"},
			expectedAppsToUpgrade:  map[apps.AppID]appOldVersion{"app1": {oldApp: app1v1, newVersion: "2"}},
			expectedAppsToRemove:   map[apps.AppID]*apps.App{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			appsToRegister, appsToUpgrade, appsToRemove := mergeApps(tc.appsForRegistration, tc.registeredApps)
			require.Equal(t, tc.expectedAppsToRegister, appsToRegister)
			require.Equal(t, tc.expectedAppsToUpgrade, appsToUpgrade)
			require.Equal(t, tc.expectedAppsToRemove, appsToRemove)
		})
	}
}
