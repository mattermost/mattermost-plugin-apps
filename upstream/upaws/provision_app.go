// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upaws

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"
)

type ProvisionAppParams struct {
	Bucket           string
	InvokePolicyName Name
	ExecuteRoleName  Name
	ShouldUpdate     bool
}

type ProvisionAppResult struct {
	InvokePolicyDoc string
	LambdaARNs      []ARN
	ManifestURL     string
	S3              []ARN
}

func ProvisionAppFromFile(c Client, path string, log Logger, params ProvisionAppParams) (*ProvisionAppResult, error) {
	provisionData, err := GetProvisionDataFromFile(path, log)
	if err != nil {
		return nil, errors.Wrapf(err, "can't get Provision data from file %s", path)
	}
	return provisionApp(c, log, provisionData, params)
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
func provisionApp(c Client, log Logger, pd *ProvisionData, params ProvisionAppParams) (*ProvisionAppResult, error) {
	var err error
	out := ProvisionAppResult{}

	err = provisionS3StaticAssets(c, log, pd, params)
	if err != nil {
		return nil, errors.Wrapf(err, "can't save manifest fo the app %s to S3", pd.Manifest.AppID)
	}

	out.LambdaARNs, out.InvokePolicyDoc, err = provisionLambdaFunctions(c, log, pd, params)
	if err != nil {
		return nil, errors.Wrapf(err, "can't provision functions of the app - %s", pd.Manifest.AppID)
	}

	err = provisionS3Manifest(c, log, pd, params)
	if err != nil {
		return nil, errors.Wrapf(err, "can't save manifest fo the app %s to S3", pd.Manifest.AppID)
	}
	return &out, nil
}

func provisionS3StaticAssets(c Client, log Logger, pd *ProvisionData, params ProvisionAppParams) error {
	for _, asset := range pd.StaticFiles {
		if err := c.UploadS3(params.Bucket, asset.Key, asset.File); err != nil {
			return errors.Wrapf(err, "failed to upload to S3: %s", asset.Key)
		}
		asset.File.Close()
		log.Info("uploaded static asset to S3.", "bucket", params.Bucket, "key", asset.Key)
	}
	return nil
}

func provisionLambdaFunctions(c Client, log Logger, pd *ProvisionData, params ProvisionAppParams) ([]ARN, string, error) {
	executeRoleARN, err := c.FindRole(params.ExecuteRoleName)
	if err != nil {
		return nil, "", err
	}
	log.Info("Found execute role, provisioning functions.", "ARN", executeRoleARN)

	createdARNs := []ARN{}
	for _, function := range pd.LambdaFunctions {
		lambdaARN := ARN("")
		if params.ShouldUpdate {
			lambdaARN, err = c.CreateOrUpdateLambda(function.Bundle, function.Name, function.Handler, function.Runtime, executeRoleARN)
			if err != nil {
				return nil, "", errors.Wrapf(err, "can't create or update function %s", function.Name)
			}
		} else {
			lambdaARN, err = c.CreateLambda(function.Bundle, function.Name, function.Handler, function.Runtime, executeRoleARN)
			if err != nil {
				return nil, "", errors.Wrapf(err, "can't create function  %s", function.Name)
			}
		}
		createdARNs = append(createdARNs, lambdaARN)
		function.Bundle.Close()
	}

	invokePolicy, err := c.FindPolicy(params.InvokePolicyName)
	if err != nil {
		return nil, "", err
	}
	invokePolicyARN := *invokePolicy.Arn
	log.Info("Found invoke policy, updating.", "ARN", invokePolicyARN)

	newDoc, err := c.AddResourcesToPolicyDocument(invokePolicy, createdARNs)
	if err != nil {
		return nil, "", err
	}

	return createdARNs, newDoc, nil
}

// provisionManifest saves manifest file in S3
func provisionS3Manifest(c Client, log Logger, pd *ProvisionData, params ProvisionAppParams) error {
	log.Info("uploading manifest to S3", "key", pd.ManifestKey)
	data, err := json.Marshal(pd.Manifest)
	if err != nil {
		return errors.Wrapf(err, "can't marshal manifest for app - %s", pd.Manifest.AppID)
	}
	buffer := bytes.NewBuffer(data)

	if err := c.UploadS3(params.Bucket, pd.ManifestKey, buffer); err != nil {
		return errors.Wrapf(err, "can't upload manifest file for the app - %s", pd.Manifest.AppID)
	}
	return nil
}
