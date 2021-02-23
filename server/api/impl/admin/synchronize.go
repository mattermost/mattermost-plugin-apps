// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package admin

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

const oldVersionKey = "update_from_version"
const appsJSONFile = "apps.json"
const callOnceKey = "CallOnce_key"

// AppVersions describes versions for all the apps in all installations
type AppVersions struct {
	Apps      apps.AppVersionMap            `json:"apps"`
	Overrides map[string]apps.AppVersionMap `json:"overrides"`
}

func getAppsForInstallation(path, installationID string) (apps.AppVersionMap, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "can't read %s file", path)
	}
	var allAppVersions *AppVersions
	if err := json.Unmarshal(data, &allAppVersions); err != nil || allAppVersions == nil {
		return nil, errors.Wrapf(err, "can't unmarshal %s file", appsJSONFile)
	}

	apps := allAppVersions.Apps
	if overrides, ok := allAppVersions.Overrides[installationID]; ok {
		for id, version := range overrides {
			apps[id] = version
		}
	}
	return apps, nil
}

func (adm *Admin) populateManifests(appVersions apps.AppVersionMap) {
	adm.store.Manifest().Cleanup()

	for id, version := range appVersions {
		manifest, err := adm.awsClient.GetManifest(id, version)
		if err != nil {
			// Note that we are not returning an error here. Instead we drop the app from the list.
			delete(appVersions, id)

			adm.mm.Log.Error("failed to get manifest", "app", id, "version", version, "err", err.Error())

			continue
		}

		adm.store.Manifest().Save(manifest)
	}
}

// LoadAppsList synchronizes apps with the apps.json file.
func (adm *Admin) LoadAppsList() error {
	bundlePath, err := adm.mm.System.GetBundlePath()
	if err != nil {
		return errors.Wrap(err, "can't get bundle path")
	}

	installationID := adm.mm.System.GetDiagnosticID()
	appsPath := filepath.Join(bundlePath, "assets", appsJSONFile)

	appsForInstallation, err := getAppsForInstallation(appsPath, installationID)
	if err != nil {
		return errors.Wrap(err, "can't get apps for installation")
	}

	adm.populateManifests(appsForInstallation)

	registeredApps := adm.store.App().GetAll()

	registeredAppsMap := map[apps.AppID]*apps.App{}
	updatedAppVersionsMap := apps.AppVersionMap{}
	// Update apps
	for _, registeredApp := range registeredApps {
		registeredAppsMap[registeredApp.Manifest.AppID] = registeredApp
		manifest, err := adm.store.Manifest().Get(registeredApp.Manifest.AppID)
		if err != nil {
			adm.mm.Log.Error("can't load manifest from store", "app_id", registeredApp.Manifest.AppID)
			continue
		}
		if err := adm.UninstallApp(registeredApp); err != nil {
			adm.mm.Log.Error("can't uninstall an app", "app_id", registeredApp.Manifest.AppID)
			continue
		}
		if manifest.Version != registeredApp.Manifest.Version {
			// update app in proxy plugin
			oldVersion := registeredApp.Manifest.Version
			registeredApp.Manifest = manifest
			if err := adm.store.App().Save(registeredApp); err != nil {
				adm.mm.Log.Error("can't save an app", "app_id", registeredApp.Manifest.AppID)
			}
			updatedAppVersionsMap[registeredApp.Manifest.AppID] = oldVersion
		}
	}

	// register new apps
	for appID, manifest := range adm.store.Manifest().GetAll() {
		if _, ok := registeredAppsMap[appID]; ok {
			continue
		}
		if err := adm.RegisterApp(manifest); err != nil {
			adm.mm.Log.Error("can't register an app", "app_id", appID)
		}
	}

	registeredAppsUpgraded := adm.store.App().GetAll()

	// call onInstanceStartup. App migration happens here
	for _, registeredApp := range registeredAppsUpgraded {
		if registeredApp.Status == apps.AppStatusInstalled {
			values := map[string]string{}
			if _, ok := updatedAppVersionsMap[registeredApp.Manifest.AppID]; ok {
				values[oldVersionKey] = string(updatedAppVersionsMap[registeredApp.Manifest.AppID])
			}
			// Call onStartup the function of the app. It should be called only once
			f := func() error {
				if err := adm.expandedCall(registeredApp, registeredApp.Manifest.OnStartup, values); err != nil {
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

func (adm *Admin) UninstallApp(app *apps.App) error {
	// Call delete the function of the app
	if err := adm.expandedCall(app, app.Manifest.OnUninstall, nil); err != nil {
		return errors.Wrapf(err, "uninstall failed. appID - %s", app.Manifest.AppID)
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

	// delete the bot account
	if err := adm.mm.Bot.DeletePermanently(app.BotUserID); err != nil {
		return errors.Wrapf(err, "can't delete bot account for App - %s", app.Manifest.AppID)
	}

	// delete app from proxy plugin, not removing the data
	if err := adm.store.App().Delete(app); err != nil {
		return errors.Wrapf(err, "can't delete app - %s", app.Manifest.AppID)
	}

	adm.mm.Log.Info("Uninstalled the app", "app_id", app.Manifest.AppID)

	return nil
}

func (adm *Admin) RegisterApp(manifest *apps.Manifest) error {
	newApp := &apps.App{}
	newApp.Manifest = manifest
	newApp.Status = apps.AppStatusRegistered
	if err := adm.store.App().Save(newApp); err != nil {
		return errors.Wrapf(err, "can't store app - %s", manifest.AppID)
	}
	adm.mm.Log.Info("App is registered", "app_id", manifest.AppID)
	return nil
}

func (adm *Admin) callOnce(f func() error) error {
	// Delete previous job
	if err := adm.mm.KV.Delete(callOnceKey); err != nil {
		return errors.Wrap(err, "can't delete key")
	}
	// Ensure all instances run this
	time.Sleep(10 * time.Second)

	adm.mutex.Lock()
	defer adm.mutex.Unlock()
	value := 0
	if err := adm.mm.KV.Get(callOnceKey, &value); err != nil {
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
	ok, err := adm.mm.KV.Set(callOnceKey, value)
	if err != nil {
		return errors.Wrapf(err, "can't set key %s to %d", callOnceKey, value)
	}
	if !ok {
		return errors.Errorf("can't set key %s to %d", callOnceKey, value)
	}
	return nil
}

func (adm *Admin) expandedCall(app *apps.App, call *apps.Call, values map[string]string) error {
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
