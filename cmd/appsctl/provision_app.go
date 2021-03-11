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
	bundle  io.Reader
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
func ProvisionApp(awscli awsclient.Client, b []byte, shouldUpdate bool) error {
	zipReader, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return errors.Wrap(err, "can't get zip reader")
	}
	bundleFunctions := []functionInstallData{}
	var m *apps.Manifest
	assets := []assetData{}

	// Read all the files from zip archive
	for _, file := range zipReader.File {
		switch {
		case strings.HasSuffix(file.Name, "manifest.json"):
			var f io.ReadCloser
			f, err = file.Open()
			if err != nil {
				return errors.Wrap(err, "can't open manifest.json file")
			}
			defer f.Close()

			var data []byte
			data, err = ioutil.ReadAll(f)
			if err != nil {
				return errors.Wrap(err, "can't read manifest.json file")
			}
			m, err = apps.ManifestFromJSON(data)
			if err != nil {
				return errors.Wrapf(err, "file %s", string(data))
			}

		case strings.HasSuffix(file.Name, ".zip"):
			var f io.ReadCloser
			f, err = file.Open()
			if err != nil {
				return errors.Wrapf(err, "can't open file %s", file.Name)
			}
			defer f.Close()

			bundleFunctions = append(bundleFunctions, functionInstallData{
				name:   strings.TrimSuffix(file.Name, ".zip"),
				bundle: f,
			})
			log.Debug("Found function bundle", "file", file.Name)

		case strings.HasPrefix(file.Name, apps.StaticAssetsFolder+"/"):
			assetName := strings.TrimPrefix(file.Name, apps.StaticAssetsFolder+"/")
			var f io.ReadCloser
			f, err = file.Open()
			if err != nil {
				return errors.Wrapf(err, "can't open file %s", file.Name)
			}
			defer f.Close()

			assets = append(assets, assetData{
				name: assetName,
				file: f,
			})
			log.Debug("Found static asset", "file", assetName)

		default:
			log.Info("Unknown file found in app bundle", "file", file.Name)
		}
	}

	if m == nil {
		return errors.New("no manifest found")
	}

	resFunctions := []functionInstallData{}

	// Matching bundle functions to the functions listed in manifest
	// O(n^2) code for simplicity
	for _, bundleFunction := range bundleFunctions {
		for _, manifestFunction := range m.AWSLambda {
			if strings.HasSuffix(bundleFunction.name, manifestFunction.Name) {
				resFunctions = append(resFunctions, functionInstallData{
					bundle:  bundleFunction.bundle,
					name:    manifestFunction.Name,
					handler: manifestFunction.Handler,
					runtime: manifestFunction.Runtime,
				})
				continue
			}
		}
	}

	if err = provisionAssets(awscli, m, assets); err != nil {
		return errors.Wrapf(err, "can't provision assets of app - %s", m.AppID)
	}
	if err = provisionFunctions(awscli, m, resFunctions, shouldUpdate); err != nil {
		return errors.Wrapf(err, "can't provision functions of app - %s", m.AppID)
	}
	if err = provisionManifest(awscli, m); err != nil {
		return errors.Wrapf(err, "can't provision manifest of app - %s", m.AppID)
	}
	return nil
}

// func ProvisionAppFromURL(awscli awsclient.Client, releaseURL string, shouldUpdate bool) error {
// 	bundle, err := httputils.GetFromURL(releaseURL)
// 	if err != nil {
// 		return errors.Wrapf(err, "can't install app from url %s", releaseURL)
// 	}

// 	return ProvisionApp(awscli, bundle, shouldUpdate)
// }

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

func provisionAssets(awscli awsclient.Client, m *apps.Manifest, assets []assetData) error {
	for _, asset := range assets {
		bucket := apps.S3BucketNameWithDefaults("")
		key := apps.AssetS3Name(m.AppID, m.Version, asset.name)
		if err := awscli.UploadS3(bucket, key, asset.file); err != nil {
			return errors.Wrapf(err, "can't provision asset - %s with key - %s", asset.name, key)
		}
	}
	return nil
}

func provisionFunctions(awscli awsclient.Client, m *apps.Manifest, functions []functionInstallData, shouldUpdate bool) error {
	policyName, err := awscli.MakeLambdaFunctionDefaultPolicy()
	if err != nil {
		return errors.Wrapf(err, "can't install app %s", m.AppID)
	}

	for _, function := range functions {
		name := apps.LambdaName(m.AppID, m.Version, function.name)

		if shouldUpdate {
			if err := awscli.CreateOrUpdateLambda(function.bundle, name, function.handler, function.runtime, policyName); err != nil {
				return errors.Wrapf(err, "can't install function for %s", m.AppID)
			}
		} else {
			if err := awscli.CreateLambda(function.bundle, name, function.handler, function.runtime, policyName); err != nil {
				return errors.Wrapf(err, "can't install function for %s", m.AppID)
			}
		}
	}

	return nil
}

func provisionManifest(awscli awsclient.Client, m *apps.Manifest) error {
	data, err := json.Marshal(m)
	if err != nil {
		return errors.Wrapf(err, "can't marshal manifest for app - %s", m.AppID)
	}
	buffer := bytes.NewBuffer(data)

	bucket := apps.S3BucketNameWithDefaults("")
	key := apps.ManifestS3Name(m.AppID, m.Version)
	if err := awscli.UploadS3(bucket, key, buffer); err != nil {
		return errors.Wrapf(err, "can't upload manifest file for the app - %s", m.AppID)
	}

	return nil
}
