// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package admin

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

// appsMappingsEnvVarName determines an environment variable.
// Variable saves address of installation's app mapping's file.
// File is given by a bucket name and key delimited by a `/`
// example: export MM_APPS_MAPPINGS = "apps/latest_mappings"
const appsMappingsEnvVarName = "MM_APPS_MAPPINGS"

const oldVersionKey = "update_from_version"

// Mappings describes mappings for all the apps in this installation
// for now it supports lambda functions and static assets
// appID -> app manifest
type Mappings struct {
	Apps map[api.AppID]*api.Manifest
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

	listedApps := adm.store.ListApps()

	listedAppsMap := map[api.AppID]*api.App{}
	updatedAppVersionsMap := map[api.AppID]string{}
	// Update apps
	for _, listedApp := range listedApps {
		listedAppsMap[listedApp.Manifest.AppID] = listedApp
		manifest, ok := mappings.Apps[listedApp.Manifest.AppID]
		if !ok {
			// TODO should deleting the app be that easy?
			if err := adm.DeleteApp(listedApp); err != nil {
				return errors.Wrapf(err, "can't delete an app")
			}
		}
		if manifest.Version != listedApp.Manifest.Version {
			// update app in proxy plugin
			oldVersion := listedApp.Manifest.Version
			listedApp.Manifest.Version = manifest.Version
			if err := adm.store.StoreApp(listedApp); err != nil {
				return errors.Wrap(err, "can't store app")
			}
			updatedAppVersionsMap[listedApp.Manifest.AppID] = oldVersion
		}
	}

	// Add new apps as listed
	for appID, manifest := range mappings.Apps {
		if _, ok := listedAppsMap[appID]; !ok {
			if err := adm.AddApp(manifest); err != nil {
				return errors.Wrap(err, "can't add new app as listed")
			}
		}
	}

	listedAppsUpgraded := adm.store.ListApps()

	// call onInstanceStartup. App migration happens here
	for _, listedApp := range listedAppsUpgraded {
		if listedApp.Status == api.AppStatusEnabled {
			values := map[string]string{}
			if _, ok := updatedAppVersionsMap[listedApp.Manifest.AppID]; ok {
				values[oldVersionKey] = updatedAppVersionsMap[listedApp.Manifest.AppID]
			}
			// Call onStartup the function of the app
			if err := adm.call(listedApp, listedApp.Manifest.OnStartup, values); err != nil {
				adm.mm.Log.Error("Can't call onStartup func of the app", "app_id", listedApp.Manifest.AppID, "err", err.Error())
			}
		}
	}

	return nil
}

func (adm *Admin) DeleteApp(app *api.App) error {
	// Call delete the function of the app
	if err := adm.call(app, app.Manifest.Delete, nil); err != nil {
		return errors.Wrap(err, "delete failed")
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

func (adm *Admin) AddApp(manifest *api.Manifest) error {
	newApp := &api.App{}
	newApp.Manifest = manifest
	newApp.Status = api.AppStatusListed
	if err := adm.store.StoreApp(newApp); err != nil {
		return errors.Wrap(err, "can't store app")
	}
	adm.mm.Log.Info("App is listed", "app_id", manifest.AppID)
	return nil
}

func (adm *Admin) call(app *api.App, call *api.Call, values map[string]string) error {
	if call == nil {
		return nil
	}

	if call.Values == nil {
		call.Values = map[string]string{}
	}
	call.Values[api.PropOAuth2ClientSecret] = app.OAuth2ClientSecret
	for k, v := range values {
		call.Values[k] = v
	}

	if call.Expand == nil {
		call.Expand = &api.Expand{}
	}
	call.Expand.App = api.ExpandAll
	call.Expand.AdminAccessToken = api.ExpandAll

	resp := adm.proxy.Call(adm.adminToken, call)
	if resp.Type == api.CallResponseTypeError {
		return errors.Wrapf(resp, "call %s failed", call.URL)
	}
	return nil
}
