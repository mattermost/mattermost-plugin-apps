package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type VersionInfo struct {
	Version string `json:"version"`
}

func (a *restapi) initPing(h *httpin.Handler) {
	h.HandleFunc(path.Ping,
		a.Ping, httpin.RequireUser).Methods(http.MethodPost)
}

func (a *restapi) Ping(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	info := a.conf.Get().GetPluginVersionInfo()
	_ = httputils.WriteJSON(w, info)
}
