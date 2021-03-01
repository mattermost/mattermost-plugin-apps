// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/awsclient"
)

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
func ProvisionAppFromFile(awscli awsclient.Client, path string, shouldUpdate bool) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "can't read file from  path %s", path)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return errors.Wrap(err, "can't read file")
	}

	return ProvisionApp(awscli, b, shouldUpdate)
}

func ProvisionApp(awscli awsclient.Client, b []byte, shouldUpdate bool) error {
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
			log.Debug("Found function bundle", "file", file.Name)
		default:
			log.Info("Unknown file found in app bundle", "file", file.Name)
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
	return provisionApp(awscli, mani.AppID, mani.Version, resFunctions, shouldUpdate)
}

func provisionApp(awscli awsclient.Client, appID apps.AppID, appVersion apps.AppVersion, functions []functionInstallData, shouldUpdate bool) error {
	policyName, err := awscli.MakeLambdaFunctionDefaultPolicy()
	if err != nil {
		return errors.Wrapf(err, "can't install app %s", appID)
	}

	for _, function := range functions {
		name, err := awsclient.MakeLambdaName(appID, appVersion, function.name)
		if err != nil {
			return errors.Wrap(err, "can't get function name")
		}

		if shouldUpdate {
			if err := awscli.CreateOrUpdateLambda(function.zipFile, name, function.handler, function.runtime, policyName); err != nil {
				return errors.Wrapf(err, "can't install function for %s", appID)
			}
		} else {
			if err := awscli.CreateLambda(function.zipFile, name, function.handler, function.runtime, policyName); err != nil {
				return errors.Wrapf(err, "can't install function for %s", appID)
			}
		}
	}

	return nil
}
