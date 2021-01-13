// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package admin

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/blang/semver"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

// appsMappingsEnvVarName determines an environment variable.
// Variable saves address of installation's app mapping's file.
// File is given by a bucket name and key delimited by a `/`
// example: export APPS_MAPPINGS = "apps/latest_mappings"
const appsMappingsEnvVarName = "APPS_MAPPINGS"

const oldVersionKey = "old_version"
const newVersionKey = "new_version"

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
			return adm.DeleteApp(app)
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
			return adm.UpgradeApp(app, mapping.AppVersion)
		}
		if incomingVersion.Compare(currentVersion) < 0 {
			return adm.DowngradeApp(app, mapping.AppVersion)
		}
	}

	return nil
}

func (adm *Admin) DeleteApp(app *api.App) error {
	// Call delete the function of the app
	delete := app.Manifest.Delete
	if delete != nil {
		if delete.Values == nil {
			delete.Values = map[string]string{}
		}
		delete.Values[api.PropOAuth2ClientSecret] = app.OAuth2ClientSecret

		if delete.Expand == nil {
			delete.Expand = &api.Expand{}
		}
		delete.Expand.App = api.ExpandAll
		delete.Expand.AdminAccessToken = api.ExpandAll

		resp := adm.proxy.Call(adm.adminToken, delete)
		if resp.Type == api.CallResponseTypeError {
			return errors.Wrap(resp, "delete failed")
		}
	}

	// delete oauth app
	conf := adm.conf.GetConfig()
	client := model.NewAPIv4Client(conf.MattermostSiteURL)
	client.SetToken(string(adm.adminToken))

	if app.OAuth2ClientID != "" {
		success, response := client.DeleteOAuthApp(app.OAuth2ClientID)
		if !success || response.StatusCode != http.StatusNoContent {
			return errors.Wrap(response.Error, "failed to delete OAuth2 App")
		}
	}

	// delete app from proxy plugin, not removing the data
	if err := adm.store.DeleteApp(app); err != nil {
		return errors.Wrap(err, "can't delete app")
	}

	adm.mm.Log.Info("Deleted the app", "app_id", app.Manifest.AppID)

	return nil
}

func expandCall(call *api.Call, app *api.App, newVersion string) *api.Call {
	if call.Values == nil {
		call.Values = map[string]string{}
	}
	call.Values[api.PropOAuth2ClientSecret] = app.OAuth2ClientSecret
	call.Values[oldVersionKey] = app.Manifest.Version
	call.Values[newVersionKey] = newVersion

	if call.Expand == nil {
		call.Expand = &api.Expand{}
	}
	call.Expand.App = api.ExpandAll
	call.Expand.AdminAccessToken = api.ExpandAll

	return call
}

func (adm *Admin) migrateApp(app *api.App, newVersion string, call *api.Call) error {
	// call the function of the app
	if call != nil {
		call = expandCall(call, app, newVersion)

		resp := adm.proxy.Call(adm.adminToken, call)
		if resp.Type == api.CallResponseTypeError {
			return errors.Wrap(resp, "migration failed")
		}
	}

	// update app in proxy plugin
	app.Manifest.Version = newVersion
	if err := adm.store.StoreApp(app); err != nil {
		return errors.Wrap(err, "can't store app")
	}

	return nil
}

func (adm *Admin) UpgradeApp(app *api.App, newVersion string) error {
	oldVersion := app.Manifest.Version
	if err := adm.migrateApp(app, newVersion, app.Manifest.Upgrade); err != nil {
		return errors.Wrap(err, "can't upgrade app")
	}
	adm.mm.Log.Info("App is Upgraded", "app_id", app.Manifest.AppID, "from", oldVersion, "to", newVersion)
	return nil
}

func (adm *Admin) DowngradeApp(app *api.App, newVersion string) error {
	oldVersion := app.Manifest.Version
	if err := adm.migrateApp(app, newVersion, app.Manifest.Downgrade); err != nil {
		return errors.Wrap(err, "can't downgrade app")
	}
	adm.mm.Log.Info("App is Downgraded", "app_id", app.Manifest.AppID, "from", oldVersion, "to", newVersion)
	return nil
}
