package upaws

import (
	"os"
	"text/template"
)

const (
	AccessEnvVar       = "MM_APPS_AWS_ACCESS_KEY"        // nolint:gosec
	SecretEnvVar       = "MM_APPS_AWS_SECRET_KEY"        // nolint:gosec
	DeployAccessEnvVar = "MM_APPS_DEPLOY_AWS_ACCESS_KEY" // nolint:gosec
	DeploySecretEnvVar = "MM_APPS_DEPLOY_AWS_SECRET_KEY" // nolint:gosec

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

const AssumeRolePolicyDocument = `{
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

const InitialInvokePolicyDocument = `{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AllowS3",
			"Effect": "Allow",
			"Action": [
				"s3:GetObject"
			],
			"Resource": [
				"arn:aws:s3:::{{.Bucket}}/manifests/*",
				"arn:aws:s3:::{{.Bucket}}/static/*"
			]
		},
		{
			"Sid": "AllowS3list",
			"Effect": "Allow",
			"Action": [
				"s3:ListBucket"
			],
			"Resource": [
				"arn:aws:s3:::{{.Bucket}}"
			]
		}
	]
}`

type PolicyDocument struct {
	Version   string
	Statement []PolicyStatement
}

type PolicyStatement struct {
	Sid      string
	Effect   string
	Action   []string
	Resource []string
}

var InvokePolicyDocumentTemplate = template.Must(template.New("InvokePolicyDocument").Parse(InitialInvokePolicyDocument))

func DefaultAllowLambdaStatement(in PolicyStatement) PolicyStatement {
	if in.Sid == "" {
		in.Sid = "AllowLambda"
	}
	if in.Effect == "" {
		in.Effect = "Allow"
	}
	if len(in.Action) == 0 {
		in.Action = []string{"lambda:InvokeFunction"}
	}
	return in
}

func Region() string {
	name := os.Getenv(RegionEnvVar)
	if name != "" {
		return name
	}
	return DefaultRegion
}
