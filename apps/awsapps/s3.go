package awsapps

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// ManifestS3Name generates key for a specific manifest in S3,
// key can be 1024 characters long.
func S3ManifestName(appID apps.AppID, version apps.AppVersion) string {
	return fmt.Sprintf("manifests/%s_%s.json", appID, version)
}

// GenerateAssetS3Name generates key for a specific asset in S3,
// key can be 1024 characters long.
func S3StaticName(appID apps.AppID, version apps.AppVersion, name string) string {
	sanitizedName := strings.ReplaceAll(name, " ", "-")
	return fmt.Sprintf("%s/%s_%s_app/%s", apps.StaticFolder, appID, version, sanitizedName)
}

func S3BucketName() string {
	name := os.Getenv(S3BucketEnvVar)
	if name != "" {
		return name
	}
	return DefaultS3Bucket
}
