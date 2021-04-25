package restapi

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
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

	cc, err := cleanUserCallContext(a.mm, actingUserID, cc)
	if err != nil {
		httputils.WriteError(w, utils.NewInvalidError(errors.Wrap(err, "invalid call context for user")))
		return
	}

	cc = a.conf.GetConfig().SetContextDefaults(cc)

	bindings, err := a.proxy.GetBindings(sessionID, actingUserID, cc)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	httputils.WriteJSON(w, bindings)
}
