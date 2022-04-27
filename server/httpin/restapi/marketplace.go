package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) initMarketplace(h *httpin.Handler) {
	h.HandleFunc(path.Marketplace, a.GetMarketplace, httpin.RequireUser).Methods(http.MethodGet)
}

func (a *restapi) GetMarketplace(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	filter := req.URL.Query().Get("filter")
	includePlugins := req.URL.Query().Get("include_plugins") != ""

	result := a.proxy.GetListedApps(r, filter, includePlugins)
	_ = httputils.WriteJSON(w, result)
}
