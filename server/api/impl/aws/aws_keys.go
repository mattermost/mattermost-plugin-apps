// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// getFunctionName generates function name for a specific app
// name can be 64 characters long.
func getFunctionName(appID apps.AppID, version apps.AppVersion, function string) (string, error) {
	if len(appID) > appIDLengthLimit {
		return "", errors.Errorf("appID %s too long, should be %d bytes", appID, appIDLengthLimit)
	}
	if len(version) > len(versionFormat) {
		return "", errors.Errorf("version %s too long, should be in %s format", version, versionFormat)
	}
	name := fmt.Sprintf("%s-%s-%s", appID, version, function)
	if len(name) <= lambdaFunctionFileNameMaxSize {
		return name, nil
	}
	functionNameLength := lambdaFunctionFileNameMaxSize - len(appID) - len(version) - 2
	hash := sha256.Sum256([]byte(name))
	hashString := hex.EncodeToString(hash[:])
	if len(hashString) > functionNameLength {
		hashString = hashString[:functionNameLength]
	}
	name = fmt.Sprintf("%s-%s-%s", appID, version, hashString)
	return name, nil
}

// getManifestFileKey generates key for a specific manifest in S3,
// key can be 1024 characters long.
func getManifestFileKey(appID apps.AppID, version apps.AppVersion) string {
	return fmt.Sprintf("manifests/%s_%s", appID, version)
}

// getAssetFileKey generates key for a specific asset in S3,
// key can be 1024 characters long.
func getAssetFileKey(appID apps.AppID, version apps.AppVersion, name string) string {
	return fmt.Sprintf("%s%s_%s_app/%s", staticAssetsFolder, appID, version, name)
}
