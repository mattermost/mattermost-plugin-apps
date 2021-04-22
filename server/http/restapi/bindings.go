package restapi

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) handleGetBindingsHTTP(w http.ResponseWriter, req *http.Request, sessionID, actingUserID string) {
	q := req.URL.Query()
	cc := &apps.Context{
		ContextFromUserAgent: apps.ContextFromUserAgent{
			TeamID:    q.Get(config.PropTeamID),
			ChannelID: q.Get(config.PropChannelID),
			PostID:    q.Get(config.PropPostID),
			UserAgent: q.Get(config.PropUserAgent),
		},
	}

	bindings, err := a.handleGetBindings(sessionID, actingUserID, cc)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	httputils.WriteJSON(w, bindings)
}

func (a *restapi) handleGetBindings(sessionID, actingUserID string, cc *apps.Context) ([]*apps.Binding, error) {
	cc, err := a.proxy.CleanUserCallContext(actingUserID, cc)
	if err != nil {
		return nil, errors.Wrap(err, "invalid call context for user")
	}

	cc = a.conf.GetConfig().SetContextDefaults(cc)

	bindings, err := a.proxy.GetBindings(sessionID, actingUserID, cc)
	if err != nil {
		return nil, err
	}

	return bindings, nil
}
