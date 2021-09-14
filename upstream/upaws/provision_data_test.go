package upaws

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func TestGetProvisionData(t *testing.T) {
	testDir, found := utils.FindDir("tests")
	require.True(t, found)

	bundlepath := filepath.Join(testDir, "test-bundle.zip")
	provisionData, err := GetProvisionDataFromFile(bundlepath, utils.NewTestLogger())
	require.NoError(t, err)
	require.Equal(t, apps.AppID("com.mattermost.servicenow"), provisionData.Manifest.AppID)
	require.Len(t, provisionData.LambdaFunctions, 4)
	require.Len(t, provisionData.Manifest.AWSLambda.Functions, 4)
	require.Equal(t, "manifests/com.mattermost.servicenow_0.1.0.json", provisionData.ManifestKey)

	for i, tc := range []struct {
		name, handler, runtime string
	}{
		{
			name:    "com-mattermost-servicenow_0-1-0_function1",
			handler: "mattermost-app-servicenow",
			runtime: "go1.x",
		},
		{
			name:    "com-mattermost-servicenow_0-1-0_function2",
			handler: "index.handler",
			runtime: "nodejs14.x",
		},
		{
			name:    "com-mattermost-servicenow_0-1-0_function-with-spaces",
			handler: "mattermost-app-servicenow",
			runtime: "go1.x",
		},
		{
			name:    "com.mattermost.servicenow_0.1.0_95f51579baba92ea2a0a3ad98c24fcbc",
			handler: "mattermost-app-servicenow",
			runtime: "go1.x",
		},
	} {
		function, ok := provisionData.LambdaFunctions[provisionData.Manifest.AWSLambda.Functions[i].Name]
		require.True(t, ok)
		require.Equal(t, tc.name, function.Name)
		require.Equal(t, tc.handler, function.Handler)
		require.Equal(t, tc.runtime, function.Runtime)
	}

	require.Len(t, provisionData.StaticFiles, 5)

	for _, tc := range []struct {
		key, value string
	}{
		{
			key:   "photo.png",
			value: "static/com.mattermost.servicenow_0.1.0_app/photo.png",
		},
		{
			key:   "some.json",
			value: "static/com.mattermost.servicenow_0.1.0_app/some.json",
		},
		{
			key:   "static file with spaces.txt",
			value: "static/com.mattermost.servicenow_0.1.0_app/static-file-with-spaces.txt",
		},
		{
			key:   "static-file-with-very-very-very-very-very-very-very-very-long-name.txt",
			value: "static/com.mattermost.servicenow_0.1.0_app/static-file-with-very-very-very-very-very-very-very-very-long-name.txt",
		},
		{
			key:   "text.txt",
			value: "static/com.mattermost.servicenow_0.1.0_app/text.txt",
		},
	} {
		asset, ok := provisionData.StaticFiles[tc.key]
		require.True(t, ok)
		require.Equal(t, tc.value, asset.Key)
	}
}
