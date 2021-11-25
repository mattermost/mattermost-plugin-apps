package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) initMarketplace(rh *httpin.Handler) {
	rh.HandleFunc(path.Marketplace, a.GetMarketplace, httpin.RequireUser).Methods(http.MethodGet)
}

func (a *restapi) GetMarketplace(req *incoming.Request, w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	includePlugins := r.URL.Query().Get("include_plugins") != ""

	result := a.proxy.GetListedApps(req, filter, includePlugins)
	_ = httputils.WriteJSON(w, result)
}
