package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/pkg/errors"
)

func (a *restapi) assetGet(w http.ResponseWriter, req *http.Request) {
	actingUserID := req.Header.Get("Mattermost-User-Id")
	if actingUserID == "" {
		httputils.WriteUnauthorizedError(w, errors.New("user not logged in"))
		return
	}

	sessionID := req.Header.Get("MM_SESSION_ID")
	if sessionID == "" {
		httputils.WriteUnauthorizedError(w, errors.New("no user session"))
		return
	}

	_, err := a.api.Mattermost.Session.Get(sessionID)
	if err != nil {
		httputils.WriteUnauthorizedError(w, err)
		return
	}
	query := req.URL.Query()

	assetName := query.Get("name")
	appID := query.Get("app_id")

	// TODO verify that request is from the correct app

	data, err := a.api.Proxy.Asset(api.AppID(appID), assetName)
	if err != nil {
		httputils.WriteBadRequestError(w, err)
		return
	}

	if _, err := w.Write(data); err != nil {
		httputils.WriteInternalServerError(w, err)
	}
}
