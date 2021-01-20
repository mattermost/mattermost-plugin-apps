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
		return errors.Wrapf(err, "can't deserialize mappings file %s/%s", bucket, key)
	}

	registeredApps := adm.store.ListApps()

	registeredAppsMap := map[api.AppID]*api.App{}
	updatedAppVersionsMap := map[api.AppID]string{}
	// Update apps
	for _, registeredApp := range registeredApps {
		registeredAppsMap[registeredApp.Manifest.AppID] = registeredApp
		manifest, ok := mappings.Apps[registeredApp.Manifest.AppID]
		if !ok {
			// TODO should deleting the app be that easy?
			if err := adm.DeleteApp(registeredApp); err != nil {
				return errors.Wrapf(err, "can't delete an app")
			}
		}
		if manifest.Version != registeredApp.Manifest.Version {
			// update app in proxy plugin
			oldVersion := registeredApp.Manifest.Version
			registeredApp.Manifest.Version = manifest.Version
			if err := adm.store.StoreApp(registeredApp); err != nil {
				return errors.Wrapf(err, "can't store app - %s", registeredApp.Manifest.AppID)
			}
			updatedAppVersionsMap[registeredApp.Manifest.AppID] = oldVersion
		}
	}

	// Add new apps as registered
	for appID, manifest := range mappings.Apps {
		if _, ok := registeredAppsMap[appID]; !ok {
			if err := adm.AddApp(manifest); err != nil {
				return errors.Wrapf(err, "can't add new app as registered")
			}
		}
	}

	registeredAppsUpgraded := adm.store.ListApps()

	// call onInstanceStartup. App migration happens here
	for _, registeredApp := range registeredAppsUpgraded {
		if registeredApp.Status == api.AppStatusEnabled {
			values := map[string]string{}
			if _, ok := updatedAppVersionsMap[registeredApp.Manifest.AppID]; ok {
				values[oldVersionKey] = updatedAppVersionsMap[registeredApp.Manifest.AppID]
			}
			// Call onStartup the function of the app
			if err := adm.call(registeredApp, registeredApp.Manifest.OnStartup, values); err != nil {
				adm.mm.Log.Error("Can't call onStartup func of the app", "app_id", registeredApp.Manifest.AppID, "err", err.Error())
			}
		}
	}

	return nil
}

func (adm *Admin) DeleteApp(app *api.App) error {
	// Call delete the function of the app
	if err := adm.call(app, app.Manifest.OnDelete, nil); err != nil {
		return errors.Wrapf(err, "delete failed. appID - %s", app.Manifest.AppID)
	}

	// delete oauth app
	conf := adm.conf.GetConfig()
	client := model.NewAPIv4Client(conf.MattermostSiteURL)
	client.SetToken(string(adm.adminToken))

	if app.OAuth2ClientID != "" {
		success, response := client.DeleteOAuthApp(app.OAuth2ClientID)
		if !success || response.StatusCode != http.StatusNoContent {
			return errors.Wrapf(response.Error, "failed to delete OAuth2 App - %s", app.Manifest.AppID)
		}
	}

	// delete app from proxy plugin, not removing the data
	if err := adm.store.DeleteApp(app); err != nil {
		return errors.Wrapf(err, "can't delete app - %s", app.Manifest.AppID)
	}

	adm.mm.Log.Info("Deleted the app", "app_id", app.Manifest.AppID)

	return nil
}

func (adm *Admin) AddApp(manifest *api.Manifest) error {
	newApp := &api.App{}
	newApp.Manifest = manifest
	newApp.Status = api.AppStatusRegistered
	if err := adm.store.StoreApp(newApp); err != nil {
		return errors.Wrapf(err, "can't store app - %s", manifest.AppID)
	}
	adm.mm.Log.Info("App is registered", "app_id", manifest.AppID)
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
