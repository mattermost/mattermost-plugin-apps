package aws

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

func TestGetProvisionData(t *testing.T) {
	testDir, found := utils.FindDir("tests")
	require.True(t, found)

	bundlepath := filepath.Join(testDir, "app_bundle_without_assets.zip")
	provisionData, err := GetProvisionDataFromFile(bundlepath, nil)
	require.NoError(t, err)
	require.Equal(t, apps.AppID("com.mattermost.servicenow"), provisionData.Manifest.AppID)
	require.Len(t, provisionData.LambdaFunctions, 1)
	require.Len(t, provisionData.Manifest.Functions, 1)
	function, ok := provisionData.LambdaFunctions[provisionData.Manifest.Functions[0].Name]
	require.True(t, ok)
	require.Equal(t, "com-mattermost-servicenow_0-1-0_go-function", function.Name)
	require.Equal(t, "mattermost-app-servicenow", function.Handler)
	require.Equal(t, "go1.x", function.Runtime)
}
