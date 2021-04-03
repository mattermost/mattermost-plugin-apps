package gateway

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (g *gateway) static(w http.ResponseWriter, req *http.Request, actingUserID, token string) {
	vars := mux.Vars(req)

	appID := vars["app_id"]
	assetName := vars["name"]
	if appID == "" {
		httputils.WriteError(w, utils.NewInvalidError("app_id not specified"))
		return
	}
	if assetName == "" {
		httputils.WriteError(w, utils.NewInvalidError("asset name not specified"))
		return
	}

	// TODO verify that request is from the correct app

	body, status, err := g.proxy.GetAsset(apps.AppID(appID), assetName)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	w.WriteHeader(status)
	if _, err := io.Copy(w, body); err != nil {
		httputils.WriteError(w, err)
		return
	}
	if err := body.Close(); err != nil {
		httputils.WriteError(w, err)
		return
	}
}
