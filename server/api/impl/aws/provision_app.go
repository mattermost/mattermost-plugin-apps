// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

const lambdaFunctionFileNameMaxSize = 64
const appIDLengthLimit = 32
const versionFormat = "v00.00.000"

type functionInstallData struct {
	zipFile io.Reader
	name    string
	handler string
	runtime string
}

// ProvisionApp gets a release URL parses the release and creates an App in AWS
// releaseURL should contain a zip with lambda functions' zip files and a `manifest.json`
// ~/my_app.zip
//  |-- manifest.json
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
	zipFile, err := downloadFile(releaseURL)
	if err != nil {
		return errors.Wrapf(err, "can't install app from url %s", releaseURL)
	}

	return c.ProvisionApp(zipFile, shouldUpdate)
}

func (c *Client) ProvisionAppFromFile(path string, shouldUpdate bool) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "can't read file from  path %s", path)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return errors.Wrap(err, "can't read file")
	}

	return c.ProvisionApp(b, shouldUpdate)
}

func (c *Client) ProvisionApp(b []byte, shouldUpdate bool) error {
	zipReader, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return errors.Wrap(err, "can't get zip reader")
	}
	bundleFunctions := []functionInstallData{}
	var mani *apps.Manifest

	// Read all the files from zip archive
	for _, file := range zipReader.File {
		switch {
		case strings.HasSuffix(file.Name, "manifest.json"):

			manifestFile, err := file.Open()
			if err != nil {
				return errors.Wrap(err, "can't open manifest.json file")
			}
			defer manifestFile.Close()

			data, err := ioutil.ReadAll(manifestFile)
			if err != nil {
				return errors.Wrap(err, "can't read manifest.json file")
			}
			err = json.Unmarshal(data, &mani)
			if err != nil {
				return errors.Wrapf(err, "can't unmarshal manifest.json file %s", string(data))
			}
		case strings.HasSuffix(file.Name, ".zip"):
			lambdaFunctionFile, err := file.Open()
			if err != nil {
				return errors.Wrapf(err, "can't open file %s", file.Name)
			}
			defer lambdaFunctionFile.Close()

			bundleFunctions = append(bundleFunctions, functionInstallData{
				name:    strings.TrimSuffix(file.Name, ".zip"),
				zipFile: lambdaFunctionFile,
			})
			c.logger.Debug("Found function bundle", "file", file.Name)
		default:
			c.logger.Info("Unknown file found in app bundle", "file", file.Name)
		}
	}

	if mani == nil {
		return errors.New("no manifest found")
	}

	resFunctions := []functionInstallData{}

	// O(n^2) code for simplicity
	for _, bundleFunction := range bundleFunctions {
		for _, manifestFunction := range mani.Functions {
			if strings.HasSuffix(bundleFunction.name, manifestFunction.Name) {
				resFunctions = append(resFunctions, functionInstallData{
					zipFile: bundleFunction.zipFile,
					name:    manifestFunction.Name,
					handler: manifestFunction.Handler,
					runtime: manifestFunction.Runtime,
				})
				continue
			}
		}
	}
	return c.provisionApp(mani.AppID, mani.Version, resFunctions, shouldUpdate)
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

func (c *Client) provisionApp(appID apps.AppID, appVersion apps.AppVersion, functions []functionInstallData, shouldUpdate bool) error {
	policyName, err := c.makeLambdaFunctionDefaultPolicy()
	if err != nil {
		return errors.Wrapf(err, "can't install app %s", appID)
	}

	for _, function := range functions {
		name, err := getFunctionName(appID, appVersion, function.name)
		if err != nil {
			return errors.Wrap(err, "can't get function name")
		}

		if shouldUpdate {
			if err := c.createOrUpdateFunction(function.zipFile, name, function.handler, function.runtime, policyName); err != nil {
				return errors.Wrapf(err, "can't install function for %s", appID)
			}
		} else {
			if err := c.createFunction(function.zipFile, name, function.handler, function.runtime, policyName); err != nil {
				return errors.Wrapf(err, "can't install function for %s", appID)
			}
		}
	}

	return nil
}

func (c *Client) createOrUpdateFunction(zipFile io.Reader, function, handler, runtime, resource string) error {
	if zipFile == nil || function == "" {
		return errors.New("you must supply a zip file, function name, handler, ARN and runtime")
	}

	contents, err := ioutil.ReadAll(zipFile)
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
		return c.createFunction(zipFile, function, handler, runtime, resource)
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
func (c *Client) createFunction(zipFile io.Reader, function, handler, runtime, resource string) error {
	if zipFile == nil || function == "" || handler == "" || resource == "" || runtime == "" {
		return errors.Errorf("you must supply a zip file, function name, handler, ARN and runtime - %p %s %s %s %s", zipFile, function, handler, resource, runtime)
	}

	contents, err := ioutil.ReadAll(zipFile)
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
	}

	c.logger.Info(fmt.Sprintf("function named %s was created with response: %v", function, result))

	return nil
}

// getFunctionName generates function name for a specific app
// name can be 64 characters long.
func getFunctionName(appID apps.AppID, version apps.AppVersion, function string) (string, error) {
	if len(appID) > appIDLengthLimit {
		return "", errors.Errorf("appID %s too long, should be %d bytes", appID, appIDLengthLimit)
	}
	if len(version) > len(versionFormat) {
		return "", errors.Errorf("version %s too long, should be in %s format", version, versionFormat)
	}

	// Sanitized any dots used in appID and version as lambda function names can not contain dots
	// While there are other non-valid characters, a dots is the most commonly used one
	sanitizedAppID := strings.ReplaceAll(string(appID), ".", "-")
	sanitizedVersion := strings.ReplaceAll(string(version), ".", "-")

	name := fmt.Sprintf("%s_%s_%s", sanitizedAppID, sanitizedVersion, function)
	if len(name) <= lambdaFunctionFileNameMaxSize {
		return name, nil
	}
	functionNameLength := lambdaFunctionFileNameMaxSize - len(sanitizedAppID) - len(sanitizedVersion) - 2
	hash := sha256.Sum256([]byte(name))
	hashString := hex.EncodeToString(hash[:])
	if len(hashString) > functionNameLength {
		hashString = hashString[:functionNameLength]
	}
	name = fmt.Sprintf("%s-%s-%s", sanitizedAppID, sanitizedVersion, hashString)
	return name, nil
}
