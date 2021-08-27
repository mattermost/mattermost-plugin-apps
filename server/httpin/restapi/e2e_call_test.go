// +build e2e

package restapi

import (
	"testing"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/stretchr/testify/require"
)

func TestCallE2E(t *testing.T) {
	th := Setup(t)
	SetupPP(th, t)
	defer th.TearDown()

	t.Run("test KV API", func(t *testing.T) {

		manifest := apps.Manifest{
			AppID:       "some_app_id",
			AppType:     apps.AppTypeHTTP,
			HomepageURL: "https://example.org",
			HTTPRootURL: "https://example.org/root",
		}

		err := manifest.IsValid()
		require.NoError(t, err)

		err = th.SystemAdminClientPP.InstallApp(manifest)
		require.NoError(t, err)

	})
}
