// main handles deployment of the plugin to a development server using the Client4 API.
package main

import (
	"net/http"
	"os"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func getMattermostClient() (*appclient.Client, error) {
	siteURL := os.Getenv("MM_SERVICESETTINGS_SITEURL")
	adminToken := os.Getenv("MM_ADMIN_TOKEN")
	if siteURL == "" || adminToken == "" {
		return nil, errors.New("MM_SERVICESETTINGS_SITEURL and MM_ADMIN_TOKEN must be set")
	}

	return appclient.NewClient("", adminToken, siteURL), nil
}

func updateMattermost(m apps.Manifest, deployType apps.DeployType, installApp bool) error {
	appClient, err := getMattermostClient()
	if err != nil {
		return err
	}

	allListed, _, err := appClient.GetListedApps("", true)
	if err != nil {
		return errors.Wrap(err, "failed to get current listed apps from Mattermost")
	}
	d := apps.Deploy{}
	for _, listed := range allListed {
		if listed.Manifest.AppID == m.AppID {
			d = listed.Manifest.Deploy
		}
	}

	// Keep the Deploy part of the stored manifest intact, just add/update the
	// new deploy type.
	m.Deploy = d.UpdateDeploy(m.Deploy, deployType)

	_, err = appClient.StoreListedApp(m)
	if err != nil {
		return errors.Wrap(err, "failed to add local manifest to Mattermost")
	}
	log.Debugw("Updated local manifest", "app_id", m.AppID, "deploy_type", deployType)

	if installApp {
		_, err = appClient.InstallApp(m.AppID, deployType)
		if err != nil {
			return errors.Wrap(err, "failed to install the app to Mattermost")
		}
		log.Debugw("Installed app to Mattermost", "app_id", m.AppID)
	}

	return nil
}

const maxManifestSize = 10 * 1024 * 1024 // 10Mb

func installPlugin(bundlePath string) (*apps.Manifest, error) {
	appClient, err := getMattermostClient()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(bundlePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open the plugin bundle")
	}
	defer f.Close()

	pluginManifest, _, err := appClient.UploadPluginForced(f)
	if err != nil {
		return nil, errors.Wrap(err, "failed to upload the plugin to Mattermost")
	}

	_, err = appClient.EnablePlugin(pluginManifest.Id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to enable plugin on Mattermost")
	}

	manifestPath := appClient.Client4.URL + "/plugins/" + pluginManifest.Id + "/manifest.json"
	resp, err := appClient.Client4.HTTPClient.Get(manifestPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get the app manifest %s", manifestPath)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to get app manifest %s: status %v", manifestPath, resp.Status)
	}

	data, err := httputils.LimitReadAll(resp.Body, maxManifestSize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the app manifest")
	}

	m, err := apps.DecodeCompatibleManifest(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse the app manifest")
	}
	return m, nil
}
