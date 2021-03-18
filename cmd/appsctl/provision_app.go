// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/awsclient"
)

// func ProvisionAppFromURL(aws awsclient.Client, releaseURL string, shouldUpdate bool) error {
// 	bundle, err := downloadFile(releaseURL)
// 	if err != nil {
// 		return errors.Wrapf(err, "can't provision app from url %s", releaseURL)
// 	}

// 	provisionData, err := getProvisionData(bundle)
// 	if err != nil {
// 		return errors.Wrapf(err, "can't get provision data for url %s", releaseURL)
// 	}

// 	return ProvisionApp(aws, provisionData, shouldUpdate)
// }

func ProvisionAppFromFile(aws awsclient.Client, path string, shouldUpdate bool) error {
	provisionData, err := GetProvisionDataFromFile(path)
	if err != nil {
		return errors.Wrapf(err, "can't get Provision data from file %s", path)
	}
	return ProvisionApp(aws, provisionData, shouldUpdate)
}

// ProvisionApp gets a release URL parses the release and creates an App in AWS
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
func ProvisionApp(aws awsclient.Client, provisionData *ProvisionData, shouldUpdate bool) error {
	bucket := apps.S3BucketNameWithDefaults("")
	// provision assets
	for _, asset := range provisionData.StaticFiles {
		if err := aws.UploadS3(bucket, asset.Key, asset.File); err != nil {
			return errors.Wrapf(err, "can't provision asset - %s of the app - %s", asset.Key, provisionData.Manifest.AppID)
		}
	}

	if err := provisionFunctions(aws, provisionData.Manifest, provisionData.LambdaFunctions, shouldUpdate); err != nil {
		return errors.Wrapf(err, "can't provision functions of the app - %s", provisionData.Manifest.AppID)
	}

	if err := provisionManifest(aws, bucket, provisionData.Manifest); err != nil {
		return errors.Wrapf(err, "can't save manifest fo the app %s to S3", provisionData.Manifest.AppID)
	}
	return nil
}

func provisionFunctions(aws awsclient.Client, manifest *apps.Manifest, functions map[string]FunctionData, shouldUpdate bool) error {
	policyName, err := aws.MakeLambdaFunctionDefaultPolicy()
	if err != nil {
		return errors.Wrap(err, "can't make lambda function default policy")
	}

	for _, function := range functions {
		if shouldUpdate {
			if err := aws.CreateOrUpdateLambda(function.Bundle, function.Name, function.Handler, function.Runtime, policyName); err != nil {
				return errors.Wrapf(err, "can't create or update function %s", function.Name)
			}
		} else {
			if err := aws.CreateLambda(function.Bundle, function.Name, function.Handler, function.Runtime, policyName); err != nil {
				return errors.Wrapf(err, "can't create function  %s", function.Name)
			}
		}
	}

	return nil
}

// provisionManifest saves manifest file in S3
func provisionManifest(aws awsclient.Client, bucket string, manifest *apps.Manifest) error {
	data, err := json.Marshal(manifest)
	if err != nil {
		return errors.Wrapf(err, "can't marshal manifest for app - %s", manifest.AppID)
	}
	buffer := bytes.NewBuffer(data)

	key := apps.ManifestS3Name(manifest.AppID, manifest.Version)
	if err := aws.UploadS3(bucket, key, buffer); err != nil {
		return errors.Wrapf(err, "can't upload manifest file for the app - %s", manifest.AppID)
	}
	return nil
}
