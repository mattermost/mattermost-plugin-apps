package helloapp

import (
	"time"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
)

const (
	appCredentialsKey = "key_app_credentials"

	helloAppPrefix = "hello_"
	dialogPrefix   = helloAppPrefix + "dialog_"

	dialogTTL = 5 * time.Minute
)

type appCredentials struct {
	BotAccessToken     string
	BotUserID          string
	OAuth2ClientID     string
	OAuth2ClientSecret string
}

func (h *helloapp) storeAppCredentials(ac *appCredentials) error {
	_, err := h.apps.Mattermost.KV.Set(appCredentialsKey, ac)
	return err
}

func (h *helloapp) getAppCredentials() (*appCredentials, error) {
	creds := appCredentials{}
	err := h.apps.Mattermost.KV.Get(appCredentialsKey, &creds)
	if err != nil {
		return nil, err
	}
	return &creds, nil
}

func (h *helloapp) storeDialog(dialog *model.OpenDialogRequest) (string, error) {
	id := model.NewId()
	_, err := h.apps.Mattermost.KV.Set(dialogPrefix+id, dialog, pluginapi.SetExpiry(dialogTTL))
	if err != nil {
		return "", err
	}

	return id, nil
}

func (h *helloapp) getDialog(dialogID string) (*model.OpenDialogRequest, error) {
	var dialog model.OpenDialogRequest
	err := h.apps.Mattermost.KV.Get(dialogPrefix+dialogID, &dialog)
	if err != nil {
		return nil, err
	}

	return &dialog, nil
}
