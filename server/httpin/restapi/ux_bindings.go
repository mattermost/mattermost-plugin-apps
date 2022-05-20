package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin/handler"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) initGetBindings(h *handler.Handler) {
	h.HandleFunc(path.Bindings,
		a.GetBindings, h.RequireActingUser).Methods(http.MethodGet)
}

// GetBindings returns combined bindings for all Apps.
//   Path: /api/v1/bindings
//   Method: GET
//   Input: none
//   Output: []Binding
func (a *restapi) GetBindings(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	bindings, err := a.Proxy.GetBindings(r, apps.Context{
		UserAgentContext: apps.UserAgentContext{
			TeamID:    q.Get(config.PropTeamID),
			ChannelID: q.Get(config.PropChannelID),
			UserAgent: q.Get(config.PropUserAgent),
		},
	})

	apiTestFlag := q.Get("test") != ""
	if apiTestFlag {
		testOut := map[string]interface{}{
			"bindings": bindings,
		}
		if err != nil {
			testOut["error"] = err.Error()
		}
		_ = httputils.WriteJSON(w, testOut)
		return
	}

	_ = httputils.WriteJSON(w, bindings)
}
