// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upaws

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/pkg/errors"
)

type ProvisionAppParams struct {
	Bucket           string
	InvokePolicyName Name
	ExecuteRoleName  Name
	ShouldUpdate     bool
}

type ProvisionAppResult struct {
	InvokePolicyDoc  string
	InvokePolicyARN  ARN
	ExecuteRoleARN   ARN
	ExecutePolicyARN ARN
	LambdaARNs       []ARN
	StaticARNs       []ARN
	ManifestURL      string
	Manifest         apps.Manifest
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
	out := ProvisionAppResult{
		Manifest: *pd.Manifest,
	}

	err := provisionS3StaticAssets(c, log, pd, params, &out)
	if err != nil {
		return nil, errors.Wrapf(err, "can't save manifest fo the app %s to S3", pd.Manifest.AppID)
	}

	err = provisionLambdaFunctions(c, log, pd, params, &out)
	if err != nil {
		return nil, errors.Wrapf(err, "can't provision functions of the app - %s", pd.Manifest.AppID)
	}

	err = provisionS3Manifest(c, log, pd, params, &out)
	if err != nil {
		return nil, errors.Wrapf(err, "can't save manifest fo the app %s to S3", pd.Manifest.AppID)
	}

	return &out, nil
}

func provisionS3StaticAssets(c Client, log Logger, pd *ProvisionData, params ProvisionAppParams, out *ProvisionAppResult) error {
	var arns []ARN
	for _, asset := range pd.StaticFiles {
		_, err := c.UploadS3(params.Bucket, asset.Key, asset.File, false)
		if err != nil {
			return errors.Wrapf(err, "failed to upload to S3: %s", asset.Key)
		}
		asset.File.Close()
		arn := ARN(fmt.Sprintf("arn:aws:s3:::%s/%s", params.Bucket, asset.Key))
		arns = append(arns, arn)
		log.Info("uploaded static asset to S3.", "bucket", params.Bucket, "key", asset.Key)
	}

	out.StaticARNs = arns
	return nil
}

func provisionLambdaFunctions(c Client, log Logger, pd *ProvisionData, params ProvisionAppParams, out *ProvisionAppResult) error {
	executeRoleARN, err := c.FindRole(params.ExecuteRoleName)
	if err != nil {
		return err
	}
	log.Info("found execute role, provisioning functions.", "ARN", executeRoleARN)

	createdARNs := []ARN{}
	for _, function := range pd.LambdaFunctions {
		lambdaARN := ARN("")
		if params.ShouldUpdate {
			lambdaARN, err = c.CreateOrUpdateLambda(function.Bundle, function.Name, function.Handler, function.Runtime, executeRoleARN)
			if err != nil {
				return errors.Wrapf(err, "can't create or update function %s", function.Name)
			}
		} else {
			lambdaARN, err = c.CreateLambda(function.Bundle, function.Name, function.Handler, function.Runtime, executeRoleARN)
			if err != nil {
				return errors.Wrapf(err, "can't create function  %s", function.Name)
			}
		}
		createdARNs = append(createdARNs, lambdaARN)
		function.Bundle.Close()
	}

	invokePolicy, err := c.FindPolicy(params.InvokePolicyName)
	if err != nil {
		return err
	}
	invokePolicyARN := ARN(*invokePolicy.Arn)
	log.Info("found invoke policy, updating.", "ARN", invokePolicyARN)

	newDoc, err := c.AddResourcesToPolicyDocument(invokePolicy, createdARNs)
	if err != nil {
		return err
	}

	out.LambdaARNs = createdARNs
	out.InvokePolicyDoc = newDoc
	out.InvokePolicyARN = invokePolicyARN
	out.ExecuteRoleARN = executeRoleARN
	out.ExecutePolicyARN = LambdaExecutionPolicyARN
	return nil
}

// provisionS3Manifest saves manifest file in S3.
func provisionS3Manifest(c Client, log Logger, pd *ProvisionData, params ProvisionAppParams, out *ProvisionAppResult) error {
	data, err := json.Marshal(pd.Manifest)
	if err != nil {
		return errors.Wrapf(err, "can't marshal manifest for app - %s", pd.Manifest.AppID)
	}
	buffer := bytes.NewBuffer(data)

	// Make the manifest publicly visible.
	url, err := c.UploadS3(params.Bucket, pd.ManifestKey, buffer, true)
	if err != nil {
		return errors.Wrapf(err, "can't upload manifest file for the app - %s", pd.Manifest.AppID)
	}

	out.Manifest = *pd.Manifest
	out.ManifestURL = url
	log.Info("uploaded manifest to S3 (public-read).", "bucket", params.Bucket, "key", pd.ManifestKey)
	return nil
}
