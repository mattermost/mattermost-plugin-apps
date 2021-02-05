// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package admin

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"
)

const oldVersionKey = "update_from_version"

// AppVersions describes versions for all the apps in all installations
type AppVersions struct {
	Apps      map[apps.AppID]string            `json:"apps"`
	Overrides map[string]map[apps.AppID]string `json:"overrides"`
}

func getAppsForInstallation(installationID string) (map[apps.AppID]string, error) {
	data, err := ioutil.ReadFile("assets/apps.json")
	if err != nil {
		return nil, errors.Wrap(err, "can't read apps.json file")
	}
	var allAppVersions *AppVersions
	if err := json.Unmarshal(data, allAppVersions); err != nil || allAppVersions == nil {
		return nil, errors.Wrap(err, "can't unmarshal apps.json file")
	}

	apps := allAppVersions.Apps
	if overrides, ok := allAppVersions.Overrides[installationID]; ok {
		for id, version := range overrides {
			apps[id] = version
		}
	}
	return apps, nil
}

func (adm *Admin) populateManifests(appVersions map[apps.AppID]string) {
	adm.store.EmptyManifests()
	for id, version := range appVersions {
		manifest, err := adm.awsClient.GetManifest(id, version)
		if err != nil {
			// Note that we are not returning an error here.
			adm.mm.Log.Error("can't get manifest for", "app", id, "version", version, "err", err)
			continue
		}
		adm.store.StoreManifest(manifest)
	}
}

// SynchronizeApps synchronizes apps with the mappings file stored in the env var.
func (adm *Admin) SynchronizeApps() error {
	installationID := adm.mm.System.GetDiagnosticID()
	appsForInstallation, err := getAppsForInstallation(installationID)
	if err != nil {
		return errors.Wrap(err, "can't get apps for installation")
	}

	adm.populateManifests(appsForInstallation)

	registeredApps := adm.store.ListApps()

	registeredAppsMap := map[apps.AppID]*apps.App{}
	updatedAppVersionsMap := map[apps.AppID]string{}
	// Update apps
	for _, registeredApp := range registeredApps {
		registeredAppsMap[registeredApp.Manifest.AppID] = registeredApp
		manifest, err := adm.store.LoadManifest(registeredApp.Manifest.AppID)
		if err != nil {
			return errors.Wrapf(err, "can't load manifest from store appID = %s", registeredApp.Manifest.AppID)
		}
		if err := adm.DeleteApp(registeredApp); err != nil {
			return errors.Wrapf(err, "can't delete an app")
		}
		if manifest.Version != registeredApp.Manifest.Version {
			// update app in proxy plugin
			oldVersion := registeredApp.Manifest.Version
			registeredApp.Manifest = manifest
			if err := adm.store.StoreApp(registeredApp); err != nil {
				return errors.Wrapf(err, "can't store app - %s", registeredApp.Manifest.AppID)
			}
			updatedAppVersionsMap[registeredApp.Manifest.AppID] = oldVersion
		}
	}

	// Add new apps as registered
	for appID, manifest := range adm.store.ListManifests() {
		if _, ok := registeredAppsMap[appID]; ok {
			continue
		}
		if err := adm.AddApp(manifest); err != nil {
			return errors.Wrapf(err, "can't add new app as registered appID - %s", manifest.AppID)
		}
	}

	registeredAppsUpgraded := adm.store.ListApps()

	// call onInstanceStartup. App migration happens here
	for _, registeredApp := range registeredAppsUpgraded {
		if registeredApp.Status == apps.AppStatusEnabled {
			values := map[string]string{}
			if _, ok := updatedAppVersionsMap[registeredApp.Manifest.AppID]; ok {
				values[oldVersionKey] = updatedAppVersionsMap[registeredApp.Manifest.AppID]
			}
			// Call onStartup the function of the app. It should be called only once
			f := func() error {
				if err := adm.call(registeredApp, registeredApp.Manifest.OnStartup, values); err != nil {
					adm.mm.Log.Error("Can't call onStartup func of the app", "app_id", registeredApp.Manifest.AppID, "err", err.Error())
				}
				return nil
			}
			if err := adm.callOnce(f); err != nil {
				adm.mm.Log.Error("Can't callOnce the onStartup func of the app", "app_id", registeredApp.Manifest.AppID, "err", err.Error())
			}
		}
	}

	return nil
}

func (adm *Admin) DeleteApp(app *apps.App) error {
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

func (adm *Admin) AddApp(manifest *apps.Manifest) error {
	newApp := &apps.App{}
	newApp.Manifest = manifest
	newApp.Status = apps.AppStatusRegistered
	if err := adm.store.StoreApp(newApp); err != nil {
		return errors.Wrapf(err, "can't store app - %s", manifest.AppID)
	}
	adm.mm.Log.Info("App is registered", "app_id", manifest.AppID)
	return nil
}

func (adm *Admin) callOnce(f func() error) error {
	// Delete previous job
	key := "PP_CallOnce_key"
	if err := adm.mm.KV.Delete(key); err != nil {
		return errors.Wrap(err, "can't delete key")
	}
	// Ensure all instances run this
	time.Sleep(10 * time.Second)

	adm.mutex.Lock()
	defer adm.mutex.Unlock()
	value := 0
	if err := adm.mm.KV.Get(key, &value); err != nil {
		return err
	}
	if value != 0 {
		// job is already run by other instance
		return nil
	}

	// job is should be run by this instance
	if err := f(); err != nil {
		return errors.Wrap(err, "can't run the job")
	}
	value = 1
	ok, err := adm.mm.KV.Set(key, value)
	if err != nil {
		return errors.Wrapf(err, "can't set key %s to %d", key, value)
	}
	if !ok {
		return errors.Errorf("can't set key %s to %d", key, value)
	}
	return nil
}

func (adm *Admin) call(app *apps.App, call *apps.Call, values map[string]string) error {
	if call == nil {
		return nil
	}

	if call.Values == nil {
		call.Values = map[string]interface{}{}
	}
	call.Values[apps.PropOAuth2ClientSecret] = app.OAuth2ClientSecret
	for k, v := range values {
		call.Values[k] = v
	}

	if call.Expand == nil {
		call.Expand = &apps.Expand{}
	}
	call.Expand.App = apps.ExpandAll
	call.Expand.AdminAccessToken = apps.ExpandAll

	resp := adm.proxy.Call(adm.adminToken, call)
	if resp.Type == apps.CallResponseTypeError {
		return errors.Wrapf(resp, "call %s failed", call.URL)
	}
	return nil
}
