package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) handleGetMarketplace(w http.ResponseWriter, req *http.Request, _ proxy.Incoming) {
	filter := req.URL.Query().Get("filter")

	result := a.proxy.GetListedApps(filter, false)
	httputils.WriteJSON(w, result)
}
