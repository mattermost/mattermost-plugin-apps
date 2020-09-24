package helloapp

import (
	"context"

	"github.com/mattermost/mattermost-plugin-api/experimental/bot/logger"
	"github.com/mattermost/mattermost-plugin-api/experimental/oauther"
	"golang.org/x/oauth2"
)

func (h *helloapp) InitOAuther() {
	h.OAuther = oauther.NewFromClient(h.mm, *h.GetOAuthConfig(), h.onConnect, logger.NewNilLogger(), oauther.OAuthURL("/hello/oauth2"), oauther.StorePrefix("hello_oauth_"))
}

func (h *helloapp) GetOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     h.OAuthClientID,
		ClientSecret: h.OAuthClientSecret,
		// <><> TODO Add scopes and maybe Endpoint
	}
}

func (h *helloapp) startOAuth(userID string) {
	payload := "Any information we want to pass throught the OAuth process"
	h.OAuther.AddPayload(userID, []byte(payload))
	url := h.OAuther.GetConnectURL()
	h.mm.Log.Debug("URL to DM user", "url", url)
}

func (h *helloapp) onConnect(userID string, token oauth2.Token, payload []byte) {
	h.mm.Log.Debug("User connected", "userID", userID, "token", token.AccessToken, "payload", string(payload))
}

func (h *helloapp) doOAuthedAction(userID string) {
	ctx := context.Background()
	t, err := h.OAuther.GetToken(userID)
	if err != nil {
		return
	}
	client := h.GetOAuthConfig().Client(ctx, t)
	client.Get("https://www.google.com")
	h.mm.Log.Debug("Doing action with token", "token", t.AccessToken)
}
