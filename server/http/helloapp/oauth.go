package helloapp

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-api/experimental/bot/logger"
	"github.com/mattermost/mattermost-plugin-api/experimental/oauther"
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-server/v5/model"
	"golang.org/x/oauth2"
)

func (h *helloapp) InitOAuther() {
	h.OAuther = oauther.NewFromClient(h.mm,
		*h.GetOAuthConfig(),
		h.finishOAuth2Connect,
		logger.NewNilLogger(),
		oauther.OAuthURL(constants.HelloAppPath+PathOAuth2),
		oauther.StorePrefix("hello_oauth_"))
}

func (h *helloapp) handleOAuth(w http.ResponseWriter, req *http.Request) {
	if h.OAuther == nil {
		http.Error(w, "OAuth not initialized", http.StatusInternalServerError)
		return
	}
	h.OAuther.ServeHTTP(w, req)
}

func (h *helloapp) GetOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     h.OAuthClientID,
		ClientSecret: h.OAuthClientSecret,
		// <><> TODO Add scopes and maybe Endpoint
	}
}

func (h *helloapp) startOAuth2Connect(userID string, callOnComplete apps.Call) (string, error) {
	state, err := json.Marshal(callOnComplete)
	if err != nil {
		return "", err
	}

	err = h.OAuther.AddPayload(userID, state)
	if err != nil {
		return "", err
	}
	return h.OAuther.GetConnectURL(), nil
}

func (h *helloapp) finishOAuth2Connect(userID string, token oauth2.Token, payload []byte) {
	call := apps.Call{}
	err := json.Unmarshal(payload, &call)
	if err != nil {
		return
	}

	// TODO 2/5 we should wrap the OAuther for the users as a "service" so that
	//  - startOAuth2Connect is a Call
	//  - payload for finish should be a Call
	//  - a Wish can check the presence of the acting user's OAuth2 token, and
	//    return Call startOAuth2Connect(itself)
	// for now hacking access to apps object and issuing the call from within
	// the app.

	call.Data.Context.AppID = AppID
	_, _ = h.apps.API.Call(call)
}

func (h *helloapp) asUser(userID string, f func(*model.Client4) error) error {
	t, err := h.OAuther.GetToken(userID)
	if err != nil {
		return err
	}
	mmClient := model.NewAPIv4Client(h.configurator.GetConfig().MattermostSiteURL)
	mmClient.SetOAuthToken(t.AccessToken)

	return f(mmClient)
}
