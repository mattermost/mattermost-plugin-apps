package restapi

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) initMarketplace(api *mux.Router) {
	api.HandleFunc(path.Marketplace,
		proxy.RequireUser(a.GetMarketplace)).Methods(http.MethodGet)
}

func (a *restapi) GetMarketplace(w http.ResponseWriter, req *http.Request, _ proxy.Incoming) {
	filter := req.URL.Query().Get("filter")
	includePlugins := req.URL.Query().Get("include_plugins") != ""

	result := a.proxy.GetListedApps(filter, includePlugins)
	httputils.WriteJSON(w, result)
}
