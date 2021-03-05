// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/pkg/errors"
)

const bundleStaticAssetsFolder = "static/"

type ProvisionData struct {
	S3BucketName    string                  `json:"s3_bucket_name"`
	StaticFiles     map[string]AssetData    `json:"static_files"`
	LambdaFunctions map[string]FunctionData `json:"lambda_functions"`
	Manifest        *apps.Manifest          `json:"-"`
}

type FunctionData struct {
	Bundle  io.Reader `json:"-"`
	Name    string    `json:"name"`
	Handler string    `json:"handler"`
	Runtime string    `json:"runtime"`
}

type AssetData struct {
	File io.Reader `json:"-"`
	Key  string    `json:"key"`
}

func (c *Client) GetProvisionDataFromFile(path string) (*ProvisionData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "can't read file from  path %s", path)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err, "can't read file")
	}

	return c.GetProvisionData(b)
}

func (c *Client) GetProvisionData(b []byte) (*ProvisionData, error) {
	bundleReader, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, errors.Wrap(err, "can't get zip reader")
	}
	bundleFunctions := []FunctionData{}
	var mani *apps.Manifest
	assets := []AssetData{}

	// Read all the files from zip archive
	for _, file := range bundleReader.File {
		switch {
		case strings.HasSuffix(file.Name, "manifest.json"):
			manifestFile, err := file.Open()
			if err != nil {
				return nil, errors.Wrap(err, "can't open manifest.json file")
			}
			defer manifestFile.Close()

			data, err := ioutil.ReadAll(manifestFile)
			if err != nil {
				return nil, errors.Wrap(err, "can't read manifest.json file")
			}
			if err := json.Unmarshal(data, &mani); err != nil {
				return nil, errors.Wrapf(err, "can't unmarshal manifest.json file %s", string(data))
			}
		case strings.HasSuffix(file.Name, ".zip"):
			lambdaFunctionFile, err := file.Open()
			if err != nil {
				return nil, errors.Wrapf(err, "can't open file %s", file.Name)
			}
			defer lambdaFunctionFile.Close()

			bundleFunctions = append(bundleFunctions, FunctionData{
				Name:   strings.TrimSuffix(file.Name, ".zip"),
				Bundle: lambdaFunctionFile,
			})
		case strings.HasPrefix(file.Name, bundleStaticAssetsFolder):
			assetName := strings.TrimPrefix(file.Name, bundleStaticAssetsFolder)
			if assetName == "" {
				continue
			}
			assetFile, err := file.Open()
			if err != nil {
				return nil, errors.Wrapf(err, "can't open file %s", file.Name)
			}
			defer assetFile.Close()

			assets = append(assets, AssetData{
				Key:  assetName,
				File: assetFile,
			})
			c.logger.Debug("Found function bundle", "file", file.Name)
		default:
			c.logger.Info("Unknown file found in app bundle", "file", file.Name)
		}
	}

	if mani == nil {
		return nil, errors.New("no manifest found")
	}

	resFunctions := []FunctionData{}

	// Matching bundle functions to the functions listed in manifest
	// O(n^2) code for simplicity
	for _, bundleFunction := range bundleFunctions {
		for _, manifestFunction := range mani.Functions {
			if strings.HasSuffix(bundleFunction.Name, manifestFunction.Name) {
				resFunctions = append(resFunctions, FunctionData{
					Bundle:  bundleFunction.Bundle,
					Name:    manifestFunction.Name,
					Handler: manifestFunction.Handler,
					Runtime: manifestFunction.Runtime,
				})
				continue
			}
		}
	}

	assetMap := generateAssetNames(mani, assets)
	functionMap, err := generateFunctionNames(mani, resFunctions)
	if err != nil {
		return nil, err
	}

	return &ProvisionData{
		S3BucketName:    c.AppsS3Bucket,
		StaticFiles:     assetMap,
		LambdaFunctions: functionMap,
		Manifest:        mani,
	}, nil
}

func generateAssetNames(manifest *apps.Manifest, assets []AssetData) map[string]AssetData {
	assetMap := make(map[string]AssetData, len(assets))
	for _, asset := range assets {
		assetMap[asset.Key] = AssetData{
			Key:  GetAssetFileKey(manifest.AppID, manifest.Version, asset.Key),
			File: asset.File,
		}
	}
	return assetMap
}

func generateFunctionNames(manifest *apps.Manifest, functions []FunctionData) (map[string]FunctionData, error) {
	functionMap := make(map[string]FunctionData, len(functions))
	for _, function := range functions {
		name, err := getFunctionName(manifest.AppID, manifest.Version, function.Name)
		if err != nil {
			return nil, errors.Wrap(err, "can't get function name")
		}
		functionMap[function.Name] = FunctionData{
			Name:    name,
			Bundle:  function.Bundle,
			Handler: function.Handler,
			Runtime: function.Runtime,
		}
	}
	return functionMap, nil
}
