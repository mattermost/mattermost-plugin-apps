package gateway

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/server/proxy/request"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (g *gateway) static(c *request.Context, w http.ResponseWriter, r *http.Request) {
	appID := appIDVar(r)
	if appID == "" {
		httputils.WriteError(w, utils.NewInvalidError("app_id not specified"))
		return
	}
	c.SetAppID(appID)

	vars := mux.Vars(r)
	if len(vars) == 0 {
		httputils.WriteError(w, utils.NewInvalidError("invalid URL format"))
		return
	}
	assetName, err := utils.CleanStaticPath(vars["name"])
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	// TODO verify that request is from the correct app

	body, status, err := g.proxy.GetStatic(c, appID, assetName)
	if err != nil {
		c.Log.WithError(err).Debugw("Failed to get asset", "asset_name", assetName)
		httputils.WriteError(w, err)
		return
	}

	copyHeader(w.Header(), r.Header)
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
