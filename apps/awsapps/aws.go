package awsapps

import "os"

const (
	AccessEnvVar          = "MM_APPS_AWS_ACCESS_KEY"           // nolint:gosec
	SecretEnvVar          = "MM_APPS_AWS_SECRET_KEY"           // nolint:gosec
	ProvisionAccessEnvVar = "MM_APPS_PROVISION_AWS_ACCESS_KEY" // nolint:gosec
	ProvisionSecretEnvVar = "MM_APPS_PROVISION_AWS_SECRET_KEY" // nolint:gosec

	DeprecatedCloudAccessEnvVar = "APPS_INVOKE_AWS_ACCESS_KEY" // nolint:gosec
	DeprecatedCloudSecretEnvVar = "APPS_INVOKE_AWS_SECRET_KEY" // nolint:gosec

	// S3BucketEnvVar is the environment variable containing the S3 bucket name
	// used to host Apps' assets.
	S3BucketEnvVar  = "MM_APPS_S3_BUCKET"
	DefaultS3Bucket = "mattermost-apps-bucket"

	RegionEnvVar  = "MM_APPS_AWS_REGION"
	DefaultRegion = "us-east-1"
)

func Region() string {
	name := os.Getenv(RegionEnvVar)
	if name != "" {
		return name
	}
	return DefaultRegion
}
