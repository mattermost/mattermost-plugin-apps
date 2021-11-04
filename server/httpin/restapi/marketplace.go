package restapi

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy/request"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) initMarketplace(api *mux.Router, c *request.Context) {
	api.Handle(path.Marketplace, request.AddContext(a.GetMarketplace, c).RequireSysadmin()).Methods(http.MethodGet)
}

func (a *restapi) GetMarketplace(_ *request.Context, w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	includePlugins := r.URL.Query().Get("include_plugins") != ""

	result := a.proxy.GetListedApps(filter, includePlugins)
	_ = httputils.WriteJSON(w, result)
}
