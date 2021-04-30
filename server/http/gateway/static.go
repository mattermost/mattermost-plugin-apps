package gateway

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (g *gateway) static(w http.ResponseWriter, req *http.Request, _, _ string) {
	appID := appIDVar(req)

	if appID == "" {
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

	body, status, err := g.proxy.GetAsset(appID, assetName)
	if err != nil {
		g.mm.Log.Debug("Failed to get asset", "app_id", appID, "asset_name", assetName, "error", err.Error())
		httputils.WriteError(w, err)
		return
	}

	copyHeader(w.Header(), req.Header)
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

func copyHeader(dst, src http.Header) {
	headerKey := "Content-Type"
	dst.Add(headerKey, src.Get(headerKey))
}
