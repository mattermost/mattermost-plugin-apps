package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) handleGetBindings(w http.ResponseWriter, req *http.Request, in proxy.Incoming) {
	q := req.URL.Query()

	bindings, err := a.proxy.GetBindings(in, apps.Context{
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
	httputils.WriteJSON(w, bindings)
}
