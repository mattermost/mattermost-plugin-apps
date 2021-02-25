package restapi

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/pkg/errors"
)

func (a *restapi) handleGetStaticAsset(w http.ResponseWriter, req *http.Request, actingUserID string) {
	vars := mux.Vars(req)

	assetName := vars["name"]
	appID := vars["app_id"]

	if appID == "" {
		httputils.WriteBadRequestError(w, errors.New("app_id not specified"))
		return
	}

	if assetName == "" {
		httputils.WriteBadRequestError(w, errors.New("asset name not specified"))
		return
	}

	// TODO verify that request is from the correct app

	resp, err := a.api.Proxy.GetAsset(apps.AppID(appID), assetName)
	if err != nil {
		httputils.WriteBadRequestError(w, err)
		return
	}

	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
