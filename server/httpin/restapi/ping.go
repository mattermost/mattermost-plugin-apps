package restapi

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type VersionInfo struct {
	Version string `json:"version"`
}

func (a *restapi) initPing(api *mux.Router, c *incoming.Request) {
	api.Handle(path.Ping,
		incoming.AddContext(a.Ping, c).RequireUser()).Methods(http.MethodPost)
}

func (a *restapi) Ping(c *incoming.Request, w http.ResponseWriter, r *http.Request) {
	info := c.Config().Get().GetPluginVersionInfo()
	_ = httputils.WriteJSON(w, info)
}
