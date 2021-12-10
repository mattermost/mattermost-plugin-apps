package restapi

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type VersionInfo struct {
	Version string `json:"version"`
}

func (a *restapi) initPing(api *mux.Router) {
	api.HandleFunc(path.Ping,
		proxy.RequireUser(a.Ping)).Methods("POST")
}

func (a *restapi) Ping(w http.ResponseWriter, req *http.Request, in proxy.Incoming) {
	info := a.conf.Get().GetPluginVersionInfo()
	_ = httputils.WriteJSON(w, info)
}
