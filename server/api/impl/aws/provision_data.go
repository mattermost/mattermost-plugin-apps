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

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

const bundleStaticAssetsFolder = "static/"

type ProvisionData struct {
	StaticFiles     []AssetData    `json:"static_files"`
	LambdaFunctions []FunctionData `json:"lambda_functions"`
	Manifest        *apps.Manifest `json:"-"`
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

func GetProvisionDataFromFile(path string, logger log) (*ProvisionData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "can't read file from  path %s", path)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err, "can't read file")
	}

	return getProvisionData(b, logger)
}

// getProvisionData takes app bundle zip as a byte slice and returns ProvisionData
func getProvisionData(b []byte, logger log) (*ProvisionData, error) {
	bundleReader, bundleErr := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if bundleErr != nil {
		return nil, errors.Wrap(bundleErr, "can't get zip reader")
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
			logger.Debug("Found function bundle", "file", file.Name)
		default:
			logger.Info("Unknown file found in app bundle", "file", file.Name)
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

	generatedAssets := generateAssetNames(mani, assets)
	generatedFunctions, err := generateFunctionNames(mani, resFunctions)
	if err != nil {
		return nil, err
	}

	return &ProvisionData{
		StaticFiles:     generatedAssets,
		LambdaFunctions: generatedFunctions,
		Manifest:        mani,
	}, nil
}

func generateAssetNames(manifest *apps.Manifest, assets []AssetData) []AssetData {
	generatedAssets := make([]AssetData, 0, len(assets))
	for _, asset := range assets {
		generatedAssets = append(generatedAssets, AssetData{
			Key:  GetAssetFileKey(manifest.AppID, manifest.Version, asset.Key),
			File: asset.File,
		})
	}
	return generatedAssets
}

func generateFunctionNames(manifest *apps.Manifest, functions []FunctionData) ([]FunctionData, error) {
	generatedFunctions := make([]FunctionData, 0, len(functions))
	for _, function := range functions {
		name, err := getFunctionName(manifest.AppID, manifest.Version, function.Name)
		if err != nil {
			return nil, errors.Wrap(err, "can't get function name")
		}
		generatedFunctions = append(generatedFunctions, FunctionData{
			Name:    name,
			Bundle:  function.Bundle,
			Handler: function.Handler,
			Runtime: function.Runtime,
		})
	}
	return generatedFunctions, nil
}
