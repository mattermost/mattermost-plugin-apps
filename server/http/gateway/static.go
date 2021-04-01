package gateway

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (g *gateway) static(w http.ResponseWriter, req *http.Request, actingUserID, token string) {
	if appIDVar(req) == "" {
		httputils.WriteError(w, utils.NewInvalidError("app_id not specified"))
		return
	}

	vars := mux.Vars(req)
	assetName := vars["name"]
	if assetName == "" {
		httputils.WriteError(w, utils.NewInvalidError("asset name not specified"))
		return
	}

	// TODO verify that request is from the correct app

	body, status, err := g.proxy.GetAsset(appIDVar(req), assetName)
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
