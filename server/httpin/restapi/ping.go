package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type VersionInfo struct {
	Version string `json:"version"`
}

func (a *restapi) handlePing(w http.ResponseWriter, req *http.Request, in proxy.Incoming) {
	info := a.conf.Get().GetPluginVersionInfo()
	httputils.WriteJSON(w, info)
}
