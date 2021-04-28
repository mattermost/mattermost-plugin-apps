package restapi

import (
	"net/http"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) handleCall(w http.ResponseWriter, req *http.Request, sessionID, actingUserID string) {
	call, err := apps.CallRequestFromJSONReader(req.Body)
	if err != nil {
		httputils.WriteError(w, utils.NewInvalidError(errors.Wrap(err, "failed to unmarshal Call request")))
		return
	}

	cc := cleanUserCallContext(a.mm, actingUserID, call.Context)
	cc = a.conf.GetConfig().SetContextDefaults(cc)

	call.Context = cc
	res := a.proxy.Call(sessionID, actingUserID, call)
	httputils.WriteJSON(w, res)
}

func cleanUserCallContext(mm *pluginapi.Client, userID string, cc *apps.Context) *apps.Context {
	return &apps.Context{
		ContextFromUserAgent: cc.ContextFromUserAgent,
	}
}
