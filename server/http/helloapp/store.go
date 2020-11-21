package helloapp

import (
	"encoding/json"
	"net/http"
	"path"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

const (
	appCredentialsKey = "key_app_credentials"
)

type appCredentials struct {
	BotAccessToken     string
	BotUserID          string
	OAuth2ClientID     string
	OAuth2ClientSecret string
}

func (h *helloapp) storeAppCredentials(ac *appCredentials) error {
	return h.asBot(
		func(mmclient *model.Client4, botUserID string) error {
			data := utils.ToJSON(ac)
			u := path.Join(
				h.apps.Configurator.GetConfig().PluginURL, apps.KVPath, appCredentialsKey)
			res, appErr := mmclient.DoApiRequest(http.MethodPut, u, data, "")
			if appErr != nil {
				return appErr
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusOK {
				return errors.Errorf("put failed, status %v", res.StatusCode)
			}
			return nil
		})
}

func (h *helloapp) getAppCredentials() (*appCredentials, error) {
	creds := appCredentials{}
	err := h.asBot(
		func(mmclient *model.Client4, botUserID string) error {
			u := path.Join(
				h.apps.Configurator.GetConfig().PluginURL, apps.KVPath, appCredentialsKey)
			res, appErr := mmclient.DoApiRequest(http.MethodGet, u, "", "")
			if appErr != nil {
				return appErr
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusOK {
				return errors.Errorf("put failed, status %v", res.StatusCode)
			}

			return json.NewDecoder(res.Body).Decode(&creds)
		})
	if err != nil {
		return nil, err
	}
	return &creds, nil
}
