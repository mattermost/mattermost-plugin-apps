package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/pkg/errors"
)

func (a *restapi) handleGetBindings(w http.ResponseWriter, req *http.Request, sessionID, actingUserID string) {
	q := req.URL.Query()

	cc := &apps.Context{
		ContextFromUserAgent: apps.ContextFromUserAgent{
			TeamID:    q.Get(config.PropTeamID),
			ChannelID: q.Get(config.PropChannelID),
			PostID:    q.Get(config.PropPostID),
			UserAgent: q.Get(config.PropUserAgent),
		},
	}

	cc, err := a.proxy.CleanUserCallContext(actingUserID, cc)
	if err != nil {
		httputils.WriteError(w, errors.Wrap(err, "invalid call context for user"))
		return
	}

	cc.ActingUserID = actingUserID
	cc.UserID = actingUserID

	cc = a.conf.GetConfig().SetContextDefaults(cc)

	bindings, err := a.proxy.GetBindings(sessionID, actingID(req), cc)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	httputils.WriteJSON(w, bindings)
}
