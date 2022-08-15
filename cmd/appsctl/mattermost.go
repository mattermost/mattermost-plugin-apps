// main handles deployment of the plugin to a development server using the Client4 API.
package main

import (
	"net/http"
	"os"
	"strings"

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

	return appclient.NewClient(adminToken, siteURL), nil
}

func updateMattermost(m apps.Manifest, deployType apps.DeployType, installApp bool) error {
	appClient, err := getMattermostClient()
	if err != nil {
		return err
	}

	// Update the listed app manifest and append the new deployment type if it's
	// not already listed.
	_, _, err = appClient.UpdateAppListing(appclient.UpdateAppListingRequest{
		Manifest:   m,
		AddDeploys: apps.DeployTypes{deployType},
	})
	if err != nil {
		return errors.Wrap(err, "failed to add local manifest to Mattermost")
	}
	log.Debugw("updated local manifest", "app_id", m.AppID, "deploy_type", deployType)

	if installApp {
		_, _, err = appClient.InstallApp(m.AppID, deployType)
		if err != nil {
			return errors.Wrap(err, "failed to install the app to Mattermost")
		}
		log.Debugw("installed app to Mattermost", "app_id", m.AppID)
	}

	return nil
}

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

	manifestPath := strings.Join([]string{
		appClient.Client4.URL,
		"plugins",
		pluginManifest.Id,
		apps.PluginAppPath,
		"manifest.json",
	}, "/")
	resp, err := appClient.Client4.HTTPClient.Get(manifestPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get the app manifest %s", manifestPath)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to get the app manifest %s: status %v", manifestPath, resp.Status)
	}

	data, err := httputils.LimitReadAll(resp.Body, apps.MaxManifestSize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the app manifest")
	}

	m, err := apps.DecodeCompatibleManifest(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse the app manifest")
	}
	return m, nil
}
