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
func (c *Client) InstallApp(releaseURL string) error {
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
	return c.installApp(mani.Name, resFunctions)
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
// Note that this url is comming from Marketplace and should be verified,
// but still it's good practice to validate here too.
func isValid(url string) bool {
	return strings.HasPrefix(url, "https://github.com/")
}

func (c *Client) installApp(appName string, functions []functionInstallData) error {
	policyName, err := c.makeLambdaFunctionDefaultPolicy()
	if err != nil {
		return errors.Wrapf(err, "can't install app %s", appName)
	}

	// check function name lengths
	for _, function := range functions {
		if len(function.name)+len(appName)+1 > lambdaFunctionFileNameMaxSize {
			return errors.Errorf("function file name %s should be less than %d", appName+"_"+function.name, lambdaFunctionFileNameMaxSize)
		}
	}
	for _, function := range functions {
		name := fmt.Sprintf("%s_%s", appName, function.name)
		if err := c.CreateFunction(function.zipFile, name, function.handler, function.runtime, policyName); err != nil {
			return errors.Wrapf(err, "can't install function for %s", appName)
		}
	}
	return nil
}
