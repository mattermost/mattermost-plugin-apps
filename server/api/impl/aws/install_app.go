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
)

const lambdaFunctionFileNameMaxSize = 64

type function struct {
	Name    string `json:"name"`
	Handler string `json:"handler"`
	Runtime string `json:"runtime"` // runtime can be detected automatically from reading the lambda function
}

type functionInstallData struct {
	zipFile io.Reader
	name    string
	handler string
	runtime string
}

// TODO tie up with the actual manifest
type manifest struct {
	AppID           string     `json:"app_id"`
	Name            string     `json:"name"`
	LambdaFunctions []function `json:"lambda_functions"`
}

// InstallApp gets a release URL parses the release and installs an App in AWS
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
func (s *Service) InstallApp(releaseURL string) error {
	zipFile, err := downloadFile(releaseURL)
	if err != nil {
		return errors.Wrapf(err, "can't install app from url %s", releaseURL)
	}
	zipReader, err := zip.NewReader(bytes.NewReader(zipFile), int64(len(zipFile)))
	if err != nil {
		return errors.Wrapf(err, "can't install app from url %s", releaseURL)
	}
	functions := []functionInstallData{}
	var mani manifest

	// Read all the files from zip archive
	for _, file := range zipReader.File {
		if file.Name == "manifest.json" {
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
		} else if strings.HasSuffix(file.Name, ".zip") {
			lambdaFunctionFile, err := file.Open()
			if err != nil {
				return errors.Wrapf(err, "can't open file %s", file.Name)
			}
			defer lambdaFunctionFile.Close()

			functions = append(functions, functionInstallData{
				name:    strings.TrimSuffix(file.Name, ".zip"),
				zipFile: lambdaFunctionFile,
			})
		}
	}
	resFunctions := []functionInstallData{}

	// O(n^2) code for simplicity
	for _, f := range functions {
		for _, manifestFunction := range mani.LambdaFunctions {
			if f.name == manifestFunction.Name {
				resFunctions = append(resFunctions, functionInstallData{
					zipFile: f.zipFile,
					name:    f.name,
					handler: manifestFunction.Handler,
					runtime: manifestFunction.Runtime,
				})
				continue
			}
		}
	}
	return s.installApp(mani.Name, resFunctions)
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

// TODO filter out nonvalid URLs. Maybe create black list to prevent SSRF attack.
// For now we will be using only urls from github.
// Note that this url is coming from Marketplace and should be verified,
// but still it's good practice to validate here too.
func isValid(url string) bool {
	return strings.HasPrefix(url, "https://github.com/")
}

func (s *Service) installApp(appName string, functions []functionInstallData) error {
	policyName, err := s.makeLambdaFunctionDefaultPolicy()
	if err != nil {
		return errors.Wrapf(err, "can't install app %s", appName)
	}

	// check function name lengths
	for _, function := range functions {
		name := getFunctionName(appName, function.name)
		if len(name) > lambdaFunctionFileNameMaxSize {
			return errors.Errorf("function file name %s should be less than %d", name, lambdaFunctionFileNameMaxSize)
		}
	}
	for _, function := range functions {
		name := getFunctionName(appName, function.name)
		if err := s.createFunction(function.zipFile, name, function.handler, function.runtime, policyName); err != nil {
			return errors.Wrapf(err, "can't install function for %s", appName)
		}
	}
	return nil
}

// CreateFunction method creates lambda function
func (s *Service) createFunction(zipFile io.Reader, function, handler, runtime, resource string) error {
	if zipFile == nil || function == "" || handler == "" || resource == "" || runtime == "" {
		return errors.Errorf("you must supply a zip file, function name, handler, ARN and runtime - %s %s %s %s %s", zipFile, function, handler, resource, runtime)
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

	result, err := s.lambda().CreateFunction(createArgs)
	if err != nil {
		if _, ok := err.(*lambda.ResourceConflictException); !ok {
			return errors.Wrapf(err, "Can't create function res = %v\n", result)
		}
	}
	s.logger.Infof("function named %s was created with result - %v", function, result)

	return nil
}

// getFunctionName generates function name for a specific app
func getFunctionName(appName, function string) string {
	return fmt.Sprintf("%s_%s", appName, function)
}
