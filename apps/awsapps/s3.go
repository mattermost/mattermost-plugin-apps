package awsapps

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

const (
	// S3BucketEnvVar is the environment variable containing the S3 bucket name
	// used to host Apps' assets.
	S3BucketEnvVar = "MM_APPS_S3_BUCKET"

	// DefaultS3Bucket is the default S3 bucket name.
	DefaultS3Bucket = "mattermost-apps-bucket"
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
