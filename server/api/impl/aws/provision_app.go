// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// ProvisionApp gets a release URL parses the release and creates an App in AWS
// releaseURL should contain a zip with lambda functions' zip files and a `manifest.json`
// ~/my_app.zip
//  |-- manifest.json
//  |-- static
//		|-- icon.png
//		|-- coolFile.txt
//  |-- my_nodejs_function.zip
//      |-- index.js
//      |-- node-modules
//          |-- async
//          |-- aws-sdk
//  |-- my_python_function.zip
//      |-- lambda_function.py
//      |-- __pycache__
//      |-- certifi/
func (c *Client) ProvisionAppFromURL(releaseURL string, shouldUpdate bool) error {
	bundle, err := downloadFile(releaseURL)
	if err != nil {
		return errors.Wrapf(err, "can't provision app from url %s", releaseURL)
	}

	provisionData, err := c.GetProvisionData(bundle)
	if err != nil {
		return errors.Wrapf(err, "can't get provision data for url %s", releaseURL)
	}

	return c.ProvisionApp(provisionData, shouldUpdate)
}

func (c *Client) ProvisionAppFromFile(path string, shouldUpdate bool) error {
	provisionData, err := c.GetProvisionDataFromFile(path)
	if err != nil {
		return errors.Wrapf(err, "can't get Provision data from file %s", path)
	}
	return c.ProvisionApp(provisionData, shouldUpdate)
}

func (c *Client) ProvisionApp(provisionData *ProvisionData, shouldUpdate bool) error {
	// provision assets
	for _, asset := range provisionData.StaticFiles {
		if err := c.S3FileUpload(asset.Key, asset.File); err != nil {
			return errors.Wrapf(err, "can't provision asset - %s of the app - %s", asset.Key, provisionData.Manifest.AppID)
		}
	}

	if err := c.provisionFunctions(provisionData.Manifest, provisionData.LambdaFunctions, shouldUpdate); err != nil {
		return errors.Wrapf(err, "can't provision functions of the app - %s", provisionData.Manifest.AppID)
	}

	if err := c.SaveManifest(provisionData.Manifest); err != nil {
		return errors.Wrapf(err, "can't save manifest fo the app %s to S3", provisionData.Manifest.AppID)
	}
	return nil
}

func downloadFile(url string) ([]byte, error) {
	if !isValid(url) {
		return nil, errors.Errorf("url %s is not valid", url)
	}
	/* #nosec G107 */
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "can't download file %s", url)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.Wrapf(err, "can't download file %s - status %d", url, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "can't read file")
	}

	return body, nil
}

// filter out nonvalid URLs. Maybe create black list to prevent SSRF attack.
// For now we will be using only urls from github.
func isValid(url string) bool {
	return strings.HasPrefix(url, "https://github.com/")
}

func (c *Client) provisionFunctions(manifest *apps.Manifest, functions map[string]FunctionData, shouldUpdate bool) error {
	policyName, err := c.makeLambdaFunctionDefaultPolicy()
	if err != nil {
		return errors.Wrap(err, "can't make lambda function default policy")
	}

	for _, function := range functions {
		if shouldUpdate {
			if err := c.createOrUpdateFunction(function.Bundle, function.Name, function.Handler, function.Runtime, policyName); err != nil {
				return errors.Wrapf(err, "can't create or update function %s", function.Name)
			}
		} else {
			if err := c.createFunction(function.Bundle, function.Name, function.Handler, function.Runtime, policyName); err != nil {
				return errors.Wrapf(err, "can't create function  %s", function.Name)
			}
		}
	}

	return nil
}

func (c *Client) createOrUpdateFunction(bundle io.Reader, function, handler, runtime, resource string) error {
	if bundle == nil || function == "" {
		return errors.New("you must supply a bundle, function name, handler, ARN and runtime")
	}

	contents, err := ioutil.ReadAll(bundle)
	if err != nil {
		return errors.Wrap(err, "could not read zip file")
	}

	_, err = c.Service().lambda.GetFunction(&lambda.GetFunctionInput{
		FunctionName: &function,
	})
	if err != nil {
		if _, ok := err.(*lambda.ResourceNotFoundException); !ok {
			return errors.Wrap(err, "Failed go get function")
		}

		// Create function if it doesn't exist
		return c.createFunction(bundle, function, handler, runtime, resource)
	}

	c.logger.Info("Updating existing function", "name", function)

	result, err := c.Service().lambda.UpdateFunctionCode(&lambda.UpdateFunctionCodeInput{
		ZipFile:      contents,
		FunctionName: &function,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to update function %v", function)
	}

	c.logger.Info(fmt.Sprintf("function named %s was updated", function), "result", result.String())

	return nil
}

// CreateFunction method creates lambda function
func (c *Client) createFunction(bundle io.Reader, function, handler, runtime, resource string) error {
	if bundle == nil || function == "" || handler == "" || resource == "" || runtime == "" {
		return errors.Errorf("you must supply a zip file, function name, handler, ARN and runtime - %p %s %s %s %s", bundle, function, handler, resource, runtime)
	}

	contents, err := ioutil.ReadAll(bundle)
	if err != nil {
		return errors.Wrap(err, "could not read zip file")
	}

	createCode := &lambda.FunctionCode{
		ZipFile: contents,
	}

	createArgs := &lambda.CreateFunctionInput{
		Code:         createCode,
		FunctionName: &function,
		Handler:      &handler,
		Role:         &resource,
		Runtime:      &runtime,
	}

	result, err := c.Service().lambda.CreateFunction(createArgs)
	if err != nil {
		if _, ok := err.(*lambda.ResourceConflictException); !ok {
			return errors.Wrapf(err, "Can't create function. Response: %v\n", result)
		}

		c.logger.Info(fmt.Sprintf("funcion %s already exists. Not updating it.", function))

		return nil
	}

	c.logger.Info(fmt.Sprintf("function %s was created with response: %v", function, result))

	return nil
}
