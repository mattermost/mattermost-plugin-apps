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

// ProvisionData contains all the necessary data for provisioning an app
type ProvisionData struct {
	// StaticFiles key is the name of the static file in the /static folder
	// Staticfiles value is the S3 Key where file should be provisioned
	StaticFiles map[string]AssetData `json:"static_files"`

	// LambdaFunctions key is the name of the lambda function zip bundle
	// LambdaFunctions value contains info for provisioning a function in the AWS.
	// LambdaFunctions value's Name field contains functions name in the AWS.
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

func GetProvisionDataFromFile(path string, log Logger) (*ProvisionData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "can't read file from  path %s", path)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err, "can't read file")
	}

	return getProvisionData(b, log)
}

// getProvisionData takes app bundle zip as a byte slice and returns ProvisionData
func getProvisionData(b []byte, log Logger) (*ProvisionData, error) {
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
			if log != nil {
				log.Debug("Found function bundle", "file", file.Name)
			}

		default:
			if log != nil {
				log.Info("Unknown file found in app bundle", "file", file.Name)
			}
		}
	}

	if mani == nil {
		return nil, errors.New("no manifest found")
	}

	resFunctions := []FunctionData{}

	// Matching bundle functions to the functions listed in manifest
	// O(n^2) code for simplicity
	for _, bundleFunction := range bundleFunctions {
		for _, manifestFunction := range mani.AWSLambda {
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
	generatedFunctions := generateFunctionNames(mani, resFunctions)

	pd := &ProvisionData{
		StaticFiles:     generatedAssets,
		LambdaFunctions: generatedFunctions,
		Manifest:        mani,
	}
	if err := pd.IsValid(); err != nil {
		return nil, errors.Wrap(err, "provision data is not valid")
	}
	return pd, nil
}

func generateAssetNames(manifest *apps.Manifest, assets []AssetData) map[string]AssetData {
	generatedAssets := make(map[string]AssetData, len(assets))
	for _, asset := range assets {
		generatedAssets[asset.Key] = AssetData{
			Key:  apps.AssetS3Name(manifest.AppID, manifest.Version, asset.Key),
			File: asset.File,
		}
	}
	return generatedAssets
}

func generateFunctionNames(manifest *apps.Manifest, functions []FunctionData) map[string]FunctionData {
	generatedFunctions := make(map[string]FunctionData, len(functions))
	for _, function := range functions {
		name := apps.LambdaName(manifest.AppID, manifest.Version, function.Name)
		generatedFunctions[function.Name] = FunctionData{
			Name:    name,
			Bundle:  function.Bundle,
			Handler: function.Handler,
			Runtime: function.Runtime,
		}
	}
	return generatedFunctions
}

func (pd *ProvisionData) IsValid() error {
	if pd.Manifest == nil {
		return errors.New("no manifest")
	}
	if err := pd.Manifest.IsValid(); err != nil {
		return err
	}

	if len(pd.Manifest.AWSLambda) != len(pd.LambdaFunctions) {
		return errors.New("different amount of functions in manifest and in the bundle")
	}

	for _, function := range pd.Manifest.AWSLambda {
		data, ok := pd.LambdaFunctions[function.Name]
		if !ok {
			return errors.Errorf("function %s was not found in the bundle", function)
		}
		if data.Handler != function.Handler {
			return errors.New("mismatched handler")
		}
		if data.Runtime != function.Runtime {
			return errors.New("mismatched runtime")
		}
	}

	return nil
}
