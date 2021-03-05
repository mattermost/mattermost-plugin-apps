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
	Apps apps.AppVersionMap `json:"apps"`
}

type appOldVersion struct {
	oldApp     *apps.App
	newVersion apps.AppVersion
}

type appNewVersion struct {
	newApp     *apps.App
	oldVersion apps.AppVersion
}

// LoadAppsList synchronizes apps with the apps.json file.
func (adm *Admin) LoadAppsList() error {
	bundlePath, err := adm.mm.System.GetBundlePath()
	if err != nil {
		return errors.Wrap(err, "can't get bundle path")
	}
	appsPath := filepath.Join(bundlePath, "assets", appsJSONFile)
	appsForRegistration, err := getAppsForRegistration(appsPath)
	if err != nil {
		return errors.Wrap(err, "can't get apps for installation")
	}
	adm.getAndStoreManifests(appsForRegistration)

	// Here are apps that are provisioned and registered on this installation.
	registeredApps := adm.store.App().GetAll()

	appsToRegister, appsToUpgrade, appsToRemove := mergeApps(appsForRegistration, registeredApps)
	adm.removeApps(appsToRemove)
	upgradedApps := adm.upgradeApps(appsToUpgrade)
	adm.registerApps(appsToRegister)

	registeredAppsUpgraded := adm.store.App().GetAll()

	// call onInstanceStartup. App migration happens here
	for _, registeredAppUpgraded := range registeredAppsUpgraded {
		if registeredAppUpgraded.Status == apps.AppStatusInstalled {
			upgradedApp, ok := upgradedApps[registeredAppUpgraded.AppID]
			values := map[string]string{}
			if ok {
				// app was upgraded send the message
				values[oldVersionKey] = string(upgradedApp.oldVersion)
			}
			adm.callOnStartupOnceWithValues(registeredAppUpgraded, values)
		}
	}

	return nil
}

func (adm *Admin) UninstallApp(appID apps.AppID) error {
	app, err := adm.store.App().Get(appID)
	if err != nil {
		return errors.Wrapf(err, "failed to get app. appID: %s", appID)
	}

	// Call delete the function of the app
	onUninstallRequest := &apps.CallRequest{Call: app.Manifest.OnUninstall}
	if err := adm.expandedCall(app, onUninstallRequest, nil); err != nil {
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

func getAppsForRegistration(path string) (apps.AppVersionMap, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "can't read %s file", path)
	}
	var apps *AppVersions
	if err := json.Unmarshal(data, &apps); err != nil || apps == nil {
		return nil, errors.Wrapf(err, "can't unmarshal %s file", appsJSONFile)
	}

	return apps.Apps, nil
}

func (adm *Admin) getAndStoreManifests(appVersions apps.AppVersionMap) {
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

// mergeApp merges two pile of apps and detects apps that should upgrade, register or remove
// appsForRegistration are apps from the apps.json file
// registeredApps are apps in the config already registered.
func mergeApps(appsForRegistration apps.AppVersionMap, registeredApps []*apps.App) (apps.AppVersionMap, map[apps.AppID]appOldVersion, map[apps.AppID]*apps.App) {
	appsToRegister := apps.AppVersionMap{}
	appsToUpgrade := map[apps.AppID]appOldVersion{}
	appsToRemove := map[apps.AppID]*apps.App{}

	allRegisteredAppsMap := map[apps.AppID]*apps.App{}
	for _, registeredApp := range registeredApps {
		allRegisteredAppsMap[registeredApp.AppID] = registeredApp
		newVersion, ok := appsForRegistration[registeredApp.AppID]
		if !ok {
			// app was registered but it is not in the registration list anymore.
			appsToRemove[registeredApp.AppID] = registeredApp
			continue
		}
		if newVersion != registeredApp.Manifest.Version {
			appsToUpgrade[registeredApp.AppID] = appOldVersion{
				oldApp:     registeredApp,
				newVersion: newVersion,
			}
		}
	}
	for id, appToRegister := range appsForRegistration {
		if _, ok := allRegisteredAppsMap[id]; !ok {
			// here is a new app
			appsToRegister[id] = appToRegister
		}
	}
	return appsToRegister, appsToUpgrade, appsToRemove
}

func (adm *Admin) removeApps(appsToRemove map[apps.AppID]*apps.App) {
	for _, appToRemove := range appsToRemove {
		switch appToRemove.Status {
		case apps.AppStatusInstalled:
			if err := adm.UninstallApp(appToRemove.AppID); err != nil {
				adm.mm.Log.Error("can't uninstall app", "app_id", appToRemove.AppID, "err", err.Error())
			}
		case apps.AppStatusRegistered:
			// delete app from proxy plugin
			if err := adm.store.App().Delete(appToRemove); err != nil {
				adm.mm.Log.Error("can't delete app from store", "app_id", appToRemove.AppID, "err", err.Error())
			}
		}
	}
}

func (adm *Admin) upgradeApps(appsToUpgrade map[apps.AppID]appOldVersion) map[apps.AppID]appNewVersion {
	upgradedApps := map[apps.AppID]appNewVersion{}
	for _, appToUpgrade := range appsToUpgrade {
		oldVersion := appToUpgrade.oldApp.Manifest.Version
		newManifest, err := adm.store.Manifest().Get(appToUpgrade.oldApp.AppID)
		if err != nil {
			adm.mm.Log.Error("can't load manifest from store", "app_id", appToUpgrade.oldApp.AppID, "err", err.Error())
			continue
		}
		if newManifest.Version != appToUpgrade.newVersion {
			adm.mm.Log.Error("versions do not match this should not happen", "app_id", appToUpgrade.oldApp.AppID)
			continue
		}
		upgradedApp := appToUpgrade.oldApp
		upgradedApp.Manifest = newManifest
		if err := adm.store.App().Save(upgradedApp); err != nil {
			adm.mm.Log.Error("can't save an app", "app_id", upgradedApp.AppID, "err", err.Error())
			continue
		}
		upgradedApps[upgradedApp.AppID] = appNewVersion{
			newApp:     upgradedApp,
			oldVersion: oldVersion,
		}
	}
	return upgradedApps
}

func (adm *Admin) registerApps(appsToRegister apps.AppVersionMap) {
	for id := range appsToRegister {
		manifest, err := adm.store.Manifest().Get(id)
		if err != nil {
			adm.mm.Log.Error("can't get manifest from store", "app_id", id, "err", err.Error())
			continue
		}
		if err := adm.registerApp(manifest); err != nil {
			adm.mm.Log.Error("can't register an app", "app_id", id, "err", err.Error())
		}
	}
}

func (adm *Admin) registerApp(manifest *apps.Manifest) error {
	newApp := &apps.App{}
	newApp.Manifest = manifest
	newApp.Status = apps.AppStatusRegistered
	if err := adm.store.App().Save(newApp); err != nil {
		return errors.Wrapf(err, "can't store app - %s", manifest.AppID)
	}
	adm.mm.Log.Info("App is registered", "app_id", manifest.AppID)
	return nil
}

func (adm *Admin) callOnStartupOnceWithValues(app *apps.App, values map[string]string) {
	// Call onStartup the function of the app. It should be called only once
	f := func() error {
		onStartupRequest := &apps.CallRequest{Call: app.Manifest.OnStartup}
		if err := adm.expandedCall(app, onStartupRequest, values); err != nil {
			adm.mm.Log.Error("Can't call onStartup func of the app", "app_id", app.Manifest.AppID, "err", err.Error())
		}
		return nil
	}
	if err := adm.callOnce(f); err != nil {
		adm.mm.Log.Error("Can't callOnce the onStartup func of the app", "app_id", app.Manifest.AppID, "err", err.Error())
	}
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

func (adm *Admin) expandedCall(app *apps.App, call *apps.CallRequest, values map[string]string) error {
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
		return errors.Wrapf(resp, "call %s failed", call.Path)
	}
	return nil
}
