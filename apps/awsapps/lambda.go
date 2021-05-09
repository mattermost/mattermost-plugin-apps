package awsapps

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

const (
	MaxLambdaName = 64

	LambdaAccessEnvVar = "MM_APPS_LAMBDA_ACCESS_KEY" // nolint:gosec
	LambdaSecretEnvVar = "MM_APPS_LAMBDA_SECRET_KEY" // nolint:gosec

	// to be deprecated
	CloudLambdaAccessEnvVar = "APPS_INVOKE_AWS_ACCESS_KEY" // nolint:gosec
	CloudLambdaSecretEnvVar = "APPS_INVOKE_AWS_SECRET_KEY" // nolint:gosec
)

// LambdaName generates function name for a specific app, name can be 64
// characters long.
func LambdaName(appID apps.AppID, version apps.AppVersion, function string) string {
	// Sanitized any dots used in appID and version as lambda function names can not contain dots
	// While there are other non-valid characters, a dots is the most commonly used one
	sanitizedAppID := strings.ReplaceAll(string(appID), ".", "-")
	sanitizedVersion := strings.ReplaceAll(string(version), ".", "-")
	sanitizedFunction := strings.ReplaceAll(function, " ", "-")

	name := fmt.Sprintf("%s_%s_%s", sanitizedAppID, sanitizedVersion, sanitizedFunction)
	if len(name) <= MaxLambdaName {
		return name
	}

	functionNameLength := MaxLambdaName - len(appID) - len(version) - 2
	hash := sha256.Sum256([]byte(name))
	hashString := hex.EncodeToString(hash[:])
	if len(hashString) > functionNameLength {
		hashString = hashString[:functionNameLength]
	}
	name = fmt.Sprintf("%s_%s_%s", appID, version, hashString)
	return name
}
