// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package awsapps

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/awsclient"
)

func ProvisionAppFromFile(c awsclient.Client, bucket, path string, shouldUpdate bool, log Logger) error {
	provisionData, err := GetProvisionDataFromFile(path, log)
	if err != nil {
		return errors.Wrapf(err, "can't get Provision data from file %s", path)
	}
	return provisionApp(c, bucket, provisionData, shouldUpdate, log)
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
func provisionApp(c awsclient.Client, bucket string, provisionData *ProvisionData, shouldUpdate bool, log Logger) error {
	if err := provisionS3StaticAssets(c, bucket, provisionData, log); err != nil {
		return errors.Wrapf(err, "can't save manifest fo the app %s to S3", provisionData.Manifest.AppID)
	}
	if err := provisionLambdaFunctions(c, provisionData, shouldUpdate, log); err != nil {
		return errors.Wrapf(err, "can't provision functions of the app - %s", provisionData.Manifest.AppID)
	}
	if err := provisionS3Manifest(c, bucket, provisionData, log); err != nil {
		return errors.Wrapf(err, "can't save manifest fo the app %s to S3", provisionData.Manifest.AppID)
	}
	return nil
}

func provisionS3StaticAssets(c awsclient.Client, bucket string, pd *ProvisionData, log Logger) error {
	for _, asset := range pd.StaticFiles {
		log.Info("uploading static asset to S3", "key", asset.Key)
		if err := c.UploadS3(bucket, asset.Key, asset.File); err != nil {
			return errors.Wrapf(err, "can't provision asset - %s of the app - %s", asset.Key, pd.Manifest.AppID)
		}
		asset.File.Close()
	}
	return nil
}

func provisionLambdaFunctions(c awsclient.Client, pd *ProvisionData, shouldUpdate bool, log Logger) error {
	log.Info("setting up IAM policy")
	policyName, err := c.EnsureLambdaRoleAndPolicy(log)
	if err != nil {
		return errors.Wrap(err, "can't make lambda function default policy")
	}

	log.Info("provisioning functions")
	for _, function := range pd.LambdaFunctions {
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
func provisionS3Manifest(c awsclient.Client, bucket string, pd *ProvisionData, log Logger) error {
	log.Info("uploading manifest to S3", "key", pd.ManifestKey)
	data, err := json.Marshal(pd.Manifest)
	if err != nil {
		return errors.Wrapf(err, "can't marshal manifest for app - %s", pd.Manifest.AppID)
	}
	buffer := bytes.NewBuffer(data)

	if err := c.UploadS3(bucket, pd.ManifestKey, buffer); err != nil {
		return errors.Wrapf(err, "can't upload manifest file for the app - %s", pd.Manifest.AppID)
	}
	return nil
}
