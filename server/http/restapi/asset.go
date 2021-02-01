package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) handleGetAsset(w http.ResponseWriter, req *http.Request, actingUserID string) {
	query := req.URL.Query()

	assetName := query.Get("name")
	appID := query.Get("app_id")

	// TODO verify that request is from the correct app

	data, err := a.api.Proxy.GetAsset(api.AppID(appID), assetName)
	if err != nil {
		httputils.WriteBadRequestError(w, err)
		return
	}

	if _, err := w.Write(data); err != nil {
		httputils.WriteInternalServerError(w, err)
	}
}
