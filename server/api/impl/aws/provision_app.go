// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

const lambdaFunctionFileNameMaxSize = 64
const appIDLengthLimit = 32
const versionFormat = "v00.00.000"
const staticAssetsFolder = "static/"

type functionInstallData struct {
	zipFile io.Reader
	name    string
	handler string
	runtime string
}

type assetData struct {
	file io.Reader
	name string
}

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
func (c *Client) ProvisionApp(releaseURL string) error {
	zipFile, zipErr := downloadFile(releaseURL)
	if zipErr != nil {
		return errors.Wrapf(zipErr, "can't install app from url %s", releaseURL)
	}
	zipReader, err := zip.NewReader(bytes.NewReader(zipFile), int64(len(zipFile)))
	if zipErr != nil {
		return errors.Wrapf(zipErr, "can't install app from url %s", releaseURL)
	}
	bundleFunctions := []functionInstallData{}
	var mani apps.Manifest
	assets := []assetData{}

	// Read all the files from zip archive
	for _, file := range zipReader.File {
		if strings.HasSuffix(file.Name, "manifest.json") { // nolint:gocritic
			manifestFile, err := file.Open()
			if err != nil {
				return errors.Wrap(err, "can't open manifest.json file")
			}
			defer manifestFile.Close()

			data, err := ioutil.ReadAll(manifestFile)
			if err != nil {
				return errors.Wrap(err, "can't read manifest.json file")
			}
			if err := json.Unmarshal(data, &mani); err != nil {
				return errors.Wrapf(err, "can't unmarshal manifest.json file %s", string(data))
			}
		} else if strings.HasSuffix(file.Name, ".zip") {
			lambdaFunctionFile, err := file.Open()
			if err != nil {
				return errors.Wrapf(err, "can't open file %s", file.Name)
			}
			defer lambdaFunctionFile.Close()

			bundleFunctions = append(bundleFunctions, functionInstallData{
				name:    strings.TrimSuffix(file.Name, ".zip"),
				zipFile: lambdaFunctionFile,
			})
		} else if strings.HasPrefix(file.Name, staticAssetsFolder) {
			assetName := strings.TrimPrefix(file.Name, staticAssetsFolder)
			assetFile, err := file.Open()
			if err != nil {
				return errors.Wrapf(err, "can't open file %s", file.Name)
			}
			defer assetFile.Close()

			assets = append(assets, assetData{
				name: assetName,
				file: assetFile,
			})
		}
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

	newManifest, err := c.provisionAssets(&mani, assets)
	if err != nil {
		return errors.Wrapf(err, "can't provision assets of the app - %s", mani.AppID)
	}

	if err := c.provisionFunctions(newManifest, resFunctions); err != nil {
		return errors.Wrapf(err, "can't provision functions of the app - %s", newManifest.AppID)
	}

	if err := c.SaveManifest(newManifest); err != nil {
		return errors.Wrap(err, "can't save manifest")
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

func (c *Client) provisionFunctions(manifest *apps.Manifest, functions []functionInstallData) error {
	policyName, err := c.makeLambdaFunctionDefaultPolicy()
	if err != nil {
		return errors.Wrapf(err, "can't install app %s", manifest.AppID)
	}

	for _, function := range functions {
		name, err := getFunctionName(manifest.AppID, manifest.Version, function.name)
		if err != nil {
			return errors.Wrap(err, "can't get function name")
		}
		if err := c.createFunction(function.zipFile, name, function.handler, function.runtime, policyName); err != nil {
			return errors.Wrapf(err, "can't install function for %s", manifest.AppID)
		}
	}
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
			return errors.Wrapf(err, "Can't create function res = %v\n", result)
		}
	}
	c.logger.Info(fmt.Sprintf("function named %s was created with result - %v", function, result))

	return nil
}

func (c *Client) provisionAssets(manifest *apps.Manifest, assets []assetData) (*apps.Manifest, error) {
	if manifest.Assets == nil {
		manifest.Assets = make([]apps.Asset, 0, len(assets))
	}
	for _, asset := range assets {
		key := getAssetFileKey(manifest.AppID, manifest.Version, asset.name)
		if err := c.S3FileUpload(key, asset.file); err != nil {
			return nil, errors.Wrapf(err, "can't provision asset - %s with key - %s", asset.name, key)
		}
		manifest.Assets = append(manifest.Assets, apps.Asset{
			Name:   asset.name,
			Type:   apps.S3Asset,
			Bucket: c.appsS3Bucket,
			Key:    key,
		})
	}
	return manifest, nil
}
