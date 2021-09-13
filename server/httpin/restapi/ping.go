package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type VersionInfo struct {
	Version string `json:"version"`
}

func (a *restapi) handlePing(w http.ResponseWriter, req *http.Request, sessionID, actingUserID string) {
	info := a.conf.Get().GetPluginVersionInfo()
	httputils.WriteJSON(w, info)
}
