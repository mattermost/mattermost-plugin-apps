// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package admin

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/blang/semver"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/pkg/errors"
)

// appsMappingsEnvVarName determines an environment variable.
// Variable saves address of installation's app mapping's file.
// File is given by a bucket name and key delimited by a `/`
// example: export APPS_MAPPINGS = "apps/latest_mappings"
const appsMappingsEnvVarName = "APPS_MAPPINGS"

// Mappings describes mappings for all the apps in this installation
// for now it supports lambda functions and static assets
// appID -> app mapping
type Mappings struct {
	Apps map[string]Mapping
}

// Mapping describes a mapping for a specific app in an installation
type Mapping struct {
	AppID      string
	AppVersion string
	Functions  []api.Function
	Assets     []api.Asset
}

// MappingsFromJSON deserializes a list of app mappings from json.
func MappingsFromJSON(data []byte) (Mappings, error) {
	var mappings Mappings
	if err := json.Unmarshal(data, &mappings); err != nil {
		return Mappings{}, err
	}
	return mappings, nil
}

// ToJSON serializes a list of app mappings to json.
func (m *Mappings) ToJSON() []byte {
	b, _ := json.Marshal(m)
	return b
}

// SynchronizeApps synchronizes apps upgrading, downgrading and deleting apps
// TODO support install as well? for force installs?
func (adm *Admin) SynchronizeApps() error {
	mappingsFile := os.Getenv(appsMappingsEnvVarName)
	if mappingsFile == "" {
		return nil
	}
	list := strings.Split(mappingsFile, "/")
	if len(list) != 2 {
		return errors.Errorf("Wrong format of an env var - %s", mappingsFile)
	}
	bucket := list[0]
	key := list[1]

	resp, err := adm.awsClient.S3FileDownload(bucket, key)
	if err != nil {
		return errors.Wrapf(err, "can't download file %s/%s", bucket, key)
	}

	mappings, err := MappingsFromJSON(resp)
	if err != nil {
		return errors.Wrap(err, "can't deserialize mappings file")
	}

	apps, _, err := adm.ListApps()
	if err != nil {
		return errors.Wrap(err, "can't get apps list")
	}

	for _, app := range apps {
		mapping, ok := mappings.Apps[string(app.Manifest.AppID)]
		if !ok {
			return adm.RemoveApp(app.Manifest.AppID)
		}
		incomingVersion, err := semver.Make(mapping.AppVersion)
		if err != nil {
			return errors.Wrapf(err, "can't parse incoming app version from the mappings file - %s", mapping.AppVersion)
		}

		currentVersion, err := semver.Make(app.Manifest.Version)
		if err != nil {
			return errors.Wrapf(err, "can't parse current app version - %s", app.Manifest.Version)
		}
		if incomingVersion.Compare(currentVersion) > 0 {
			return adm.UpgradeApp(app.Manifest.AppID, mapping.AppVersion)
		}
		if incomingVersion.Compare(currentVersion) < 0 {
			return adm.DowngradeApp(app.Manifest.AppID, mapping.AppVersion)
		}
	}

	return nil
}

func (adm *Admin) RemoveApp(appID api.AppID) error {
	return nil
}

func (adm *Admin) UpgradeApp(appID api.AppID, newVersion string) error {
	return nil
}

func (adm *Admin) DowngradeApp(appID api.AppID, newVersion string) error {
	return nil
}
