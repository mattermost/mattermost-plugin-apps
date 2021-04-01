package gateway

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (g *gateway) handleWebhook(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	appID := vars["app_id"]
	if appID == "" {
		httputils.WriteError(w, utils.NewInvalidError("app_id not specified"))
		return
	}

	path := vars["path"]
	if path == "" {
		httputils.WriteError(w, utils.NewInvalidError("webhook call path not specified"))
		return
	}

	secret := req.URL.Query().Get("secret")
	if secret == "" {
		httputils.WriteError(w, utils.NewInvalidError("webhook secret was not provided"))
		return
	}

	if !g.isValidSecret(appID, secret) {
		httputils.WriteError(w, utils.NewInvalidError("webhook secret is not valid"))
		return
	}

	var c interface{}
	err := json.NewDecoder(req.Body).Decode(&c)
	if err != nil {
		httputils.WriteError(w, utils.NewInvalidError(err))
		return
	}

	values := make(map[string]interface{})
	values["payload"] = c

	call := apps.CallRequest{
		Values: values,
		Context: &apps.Context{
			AppID: apps.AppID(appID),
		},
		Type: apps.CallTypeSubmit,
		Call: apps.Call{
			Path: "/" + path,
		},
	}
	_ = g.proxy.Call("", "", &call)
}

func (g *gateway) isValidSecret(appID, secret string) bool {
	app, _ := g.proxy.GetInstalledApp(apps.AppID(appID))
	savedSecret := app.WebhookSecret
	return secret == savedSecret
}
