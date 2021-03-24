package gateway

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (g *gateway) handleGetStaticAsset(w http.ResponseWriter, req *http.Request, actingUserID, token string) {
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

	body, status, err := g.proxy.GetAsset(apps.AppID(appID), assetName)
	if err != nil {
		httputils.WriteBadRequestError(w, err)
		return
	}

	w.WriteHeader(status)
	if _, err := io.Copy(w, body); err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}
	if err := body.Close(); err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}
}
