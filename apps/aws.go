package apps

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// Where static assets are.
const StaticAssetsFolder = "static"

const (
	AWSMaxLambdaName = 64

	// appsS3BucketEnvVarName determines an environment variable.
	// Variable saves address of apps S3 bucket name
	AWSS3BucketEnvVar = "MM_APPS_S3_BUCKET"

	// defaultBucketName is the default s3 bucket name used to store app data.
	AWSDefaultS3Bucket = "mattermost-apps-bucket"
)

// AWSLambdaFunction describes a distinct AWS Lambda function defined by the
// app, and what path should be mapped to it.
//
// cmd/appsctl will create or update the manifest's aws_lambda functions in the
// AWS Lambda service.
//
// upawslambda will use the manifest's aws_lambda functions to find the closest
// match for the call's path, and then to invoke the AWS Lambda function.
type AWSLambdaFunction struct {
	// The lambda function with its Path the longest-matching prefix of the
	// call's Path will be invoked for a call.
	Path string `json:"path"`

	// TODO @iomodo
	Name    string `json:"name"`
	Handler string `json:"handler"`
	Runtime string `json:"runtime"`
}

func (f AWSLambdaFunction) IsValid() error {
	if f.Path == "" {
		return errors.New("aws_lambda path must not be empty")
	}
	if f.Name == "" {
		return errors.New("aws_lambda name must not be empty")
	}
	if f.Runtime == "" {
		return errors.New("aws_lambda runtime must not be empty")
	}
	if f.Handler == "" {
		return errors.New("aws_lambda handler must not be empty")
	}
	return nil
}

// LambdaName generates function name for a specific app, name can be 64
// characters long.
func LambdaName(appID AppID, version AppVersion, function string) string {
	// Sanitized any dots used in appID and version as lambda function names can not contain dots
	// While there are other non-valid characters, a dots is the most commonly used one
	sanitizedAppID := strings.ReplaceAll(string(appID), ".", "-")
	sanitizedVersion := strings.ReplaceAll(string(version), ".", "-")

	name := fmt.Sprintf("%s_%s_%s", sanitizedAppID, sanitizedVersion, function)
	if len(name) <= AWSMaxLambdaName {
		return name
	}

	functionNameLength := AWSMaxLambdaName - len(appID) - len(version) - 2
	hash := sha256.Sum256([]byte(name))
	hashString := hex.EncodeToString(hash[:])
	if len(hashString) > functionNameLength {
		hashString = hashString[:functionNameLength]
	}
	name = fmt.Sprintf("%s_%s_%s", appID, version, hashString)
	return name
}

// ManifestS3Name generates key for a specific manifest in S3,
// key can be 1024 characters long.
func ManifestS3Name(appID AppID, version AppVersion) string {
	return fmt.Sprintf("manifests/%s_%s.json", appID, version)
}

// GenerateAssetS3Name generates key for a specific asset in S3,
// key can be 1024 characters long.
func AssetS3Name(appID AppID, version AppVersion, name string) string {
	sanitizedName := strings.ReplaceAll(name, " ", "-")
	return fmt.Sprintf("%s/%s_%s_app/%s", StaticAssetsFolder, appID, version, sanitizedName)
}

func S3BucketNameWithDefaults(name string) string {
	if name != "" {
		return name
	}
	name = os.Getenv(AWSS3BucketEnvVar)
	if name != "" {
		return name
	}
	return AWSDefaultS3Bucket
}
