package helloapp

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
