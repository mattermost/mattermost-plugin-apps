// main handles deployment of the plugin to a development server using the Client4 API.
package main

import (
	"os"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
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
