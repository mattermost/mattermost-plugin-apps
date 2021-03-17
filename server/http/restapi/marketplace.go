package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) handleGetMarketplace(w http.ResponseWriter, req *http.Request, actingUserID string) {
	filter := req.URL.Query().Get("filter")

	result := a.admin.GetListedApps(filter)
	httputils.WriteJSON(w, result)
}
