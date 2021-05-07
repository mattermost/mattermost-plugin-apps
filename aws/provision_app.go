// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (c *client) ProvisionAppFromFile(path string, shouldUpdate bool) error {
	provisionData, err := GetProvisionDataFromFile(path, c.logger)
	if err != nil {
		return errors.Wrapf(err, "can't get Provision data from file %s", path)
	}
	return c.provisionApp(provisionData, shouldUpdate)
}

// provisionApp gets a release URL parses the release and creates an App in AWS
// releaseURL should contain a zip with lambda functions' zip files and a `manifest.json`
//  ~/my_app.zip
//   |-- manifest.json
//   |-- static
//		|-- icon.png
//		|-- coolFile.txt
//   |-- my_nodejs_function.zip
//      |-- index.js
//      |-- node-modules
//          |-- async
//          |-- aws-sdk
//   |-- my_python_function.zip
//      |-- lambda_function.py
//      |-- __pycache__
//      |-- certifi/
func (c *client) provisionApp(provisionData *ProvisionData, shouldUpdate bool) error {
	bucket := apps.S3BucketNameWithDefaults("")
	// provision assets
	for _, asset := range provisionData.StaticFiles {
		if err := c.UploadS3(bucket, asset.Key, asset.File); err != nil {
			return errors.Wrapf(err, "can't provision asset - %s of the app - %s", asset.Key, provisionData.Manifest.AppID)
		}
		asset.File.Close()
	}

	if err := c.provisionFunctions(provisionData.Manifest, provisionData.LambdaFunctions, shouldUpdate); err != nil {
		return errors.Wrapf(err, "can't provision functions of the app - %s", provisionData.Manifest.AppID)
	}

	if err := c.provisionManifest(bucket, provisionData.Manifest, provisionData.ManifestKey); err != nil {
		return errors.Wrapf(err, "can't save manifest fo the app %s to S3", provisionData.Manifest.AppID)
	}
	return nil
}

func (c *client) provisionFunctions(manifest *apps.Manifest, functions map[string]FunctionData, shouldUpdate bool) error {
	policyName, err := c.MakeLambdaFunctionDefaultPolicy()
	if err != nil {
		return errors.Wrap(err, "can't make lambda function default policy")
	}

	for _, function := range functions {
		if shouldUpdate {
			if err := c.CreateOrUpdateLambda(function.Bundle, function.Name, function.Handler, function.Runtime, policyName); err != nil {
				return errors.Wrapf(err, "can't create or update function %s", function.Name)
			}
		} else {
			if err := c.CreateLambda(function.Bundle, function.Name, function.Handler, function.Runtime, policyName); err != nil {
				return errors.Wrapf(err, "can't create function  %s", function.Name)
			}
		}
		function.Bundle.Close()
	}

	return nil
}

// provisionManifest saves manifest file in S3
func (c *client) provisionManifest(bucket string, manifest *apps.Manifest, key string) error {
	data, err := json.Marshal(manifest)
	if err != nil {
		return errors.Wrapf(err, "can't marshal manifest for app - %s", manifest.AppID)
	}
	buffer := bytes.NewBuffer(data)

	if err := c.UploadS3(bucket, key, buffer); err != nil {
		return errors.Wrapf(err, "can't upload manifest file for the app - %s", manifest.AppID)
	}
	return nil
}
