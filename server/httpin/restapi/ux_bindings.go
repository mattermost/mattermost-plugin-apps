package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) initGetBindings(rh *httpin.Handler) {
	rh.HandleFunc(path.Bindings,
		a.GetBindings, httpin.RequireUser).Methods(http.MethodGet)
}

// GetBindings returns combined bindings for all Apps.
//   Path: /api/v1/bindings
//   Method: GET
//   Input: none
//   Output: []Binding
func (a *restapi) GetBindings(req *incoming.Request, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	bindings, err := a.proxy.GetBindings(req, apps.Context{
		UserAgentContext: apps.UserAgentContext{
			TeamID:    q.Get(config.PropTeamID),
			ChannelID: q.Get(config.PropChannelID),
			PostID:    q.Get(config.PropPostID),
			UserAgent: q.Get(config.PropUserAgent),
		},
	})
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	_ = httputils.WriteJSON(w, bindings)
}
