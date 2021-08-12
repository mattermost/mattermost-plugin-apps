package gateway

import (
	"crypto/subtle"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (g *gateway) handleWebhook(w http.ResponseWriter, req *http.Request) {
	appID := appIDVar(req)
	if appID == "" {
		httputils.WriteError(w, utils.NewInvalidError("app_id not specified"))
		return
	}

	secret := req.URL.Query().Get("secret")
	if secret == "" {
		httputils.WriteError(w, utils.NewInvalidError("webhook secret was not provided"))
		return
	}
	app, err := g.proxy.GetInstalledApp(appID)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	if subtle.ConstantTimeCompare([]byte(secret), []byte(app.WebhookSecret)) != 1 {
		httputils.WriteError(w, utils.NewInvalidError("webhook secret mismatched"))
		return
	}

	conf := g.conf.Get()
	data, err := httputils.LimitReadAll(req.Body, conf.MaxWebhookSize)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	vars := mux.Vars(req)
	path := vars["path"]
	if path == "" {
		httputils.WriteError(w, utils.NewInvalidError("webhook call path not specified"))
		return
	}

	_ = g.proxy.NotifyRemoteWebhook(app, data, path)
}
