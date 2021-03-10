// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package awsclient

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

const MaxLambdaName = 64
const StaticAssetsFolder = "static/"

// GenerateLambdaName generates function name for a specific app, name can be 64
// characters long.
func GenerateLambdaName(appID apps.AppID, version apps.AppVersion, function string) string {
	// Sanitized any dots used in appID and version as lambda function names can not contain dots
	// While there are other non-valid characters, a dots is the most commonly used one
	sanitizedAppID := strings.ReplaceAll(string(appID), ".", "-")
	sanitizedVersion := strings.ReplaceAll(string(version), ".", "-")

	name := fmt.Sprintf("%s_%s_%s", sanitizedAppID, sanitizedVersion, function)
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

// GenerateManifestS3Name generates key for a specific manifest in S3,
// key can be 1024 characters long.
func GenerateManifestS3Name(appID apps.AppID, version apps.AppVersion) string {
	return fmt.Sprintf("manifests/%s_%s.json", appID, version)
}

// GenerateAssetS3Name generates key for a specific asset in S3,
// key can be 1024 characters long.
func GenerateAssetS3Name(appID apps.AppID, version apps.AppVersion, name string) string {
	return fmt.Sprintf("%s%s_%s_app/%s", StaticAssetsFolder, appID, version, name)
}

func GenerateS3BucketNameWithDefaults(name string) string {
	if name != "" {
		return name
	}
	name = os.Getenv(AppsS3BucketEnvVarName)
	if name != "" {
		return name
	}
	return DefaultBucketName
}
