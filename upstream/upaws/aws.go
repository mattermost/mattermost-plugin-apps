package upaws

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

const LambdaExecutionPolicyARN = ARN(`arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole`)

const (
	DefaultExecuteRoleName = "mattermost-apps-execute-lambda-role"
	DefaultPolicyName      = "mattermost-apps-invoke-policy"
	DefaultUserName        = "mattermost-apps-invoke"
	DefaultGroupName       = "mattermost-apps-invoke-group"
)

const ExecuteRolePolicyDocument = `{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Principal": {
				"Service": "lambda.amazonaws.com"
			},
			"Action": "sts:AssumeRole"
		}
	]
}`

func Region() string {
	name := os.Getenv(RegionEnvVar)
	if name != "" {
		return name
	}
	return DefaultRegion
}
