package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) handleWebhook(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	appID := vars["app_id"]
	if appID == "" {
		httputils.WriteBadRequestError(w, errors.New("app_id not specified"))
		return
	}

	path := vars["path"]
	if path == "" {
		httputils.WriteBadRequestError(w, errors.New("webhook call path not specified"))
		return
	}

	queryVars := req.URL.Query()
	if len(queryVars["secret"]) != 1 {
		httputils.WriteBadRequestError(w, errors.New("webhook secret was not provided"))
		return
	}

	secret := queryVars["secret"][0]
	if !a.isValidSecret(appID, secret) {
		httputils.WriteBadRequestError(w, errors.New("webhook secret is not valid"))
		return
	}

	var c interface{}
	err := json.NewDecoder(req.Body).Decode(&c)
	if err != nil {
		httputils.WriteBadRequestError(w, err)
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
	_ = a.proxy.Call("", &call)
}

func (a *restapi) isValidSecret(appID, secret string) bool {
	app, _ := a.proxy.GetInstalledApp(apps.AppID(appID))
	savedSecret := app.WebhookSecret
	return secret == savedSecret
}
