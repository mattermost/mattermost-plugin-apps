package restapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/pkg/errors"
)

func (a *restapi) handleGetBindings(w http.ResponseWriter, req *http.Request, actingUserID, token string) {
	q := req.URL.Query()
	cc := a.conf.GetConfig().SetContextDefaults(&apps.Context{
		ActingUserID: actingUserID,
		TeamID:       q.Get(config.PropTeamID),
		ChannelID:    q.Get(config.PropChannelID),
		PostID:       q.Get(config.PropPostID),
		UserAgent:    q.Get(config.PropUserAgent),
		UserID:       actingUserID,
	})

	bindings, err := a.proxy.GetBindings(sessionID(req), actingID(req), cc)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	httputils.WriteJSON(w, bindings)
}

func (a *restapi) handleCacheInvalidateBindings(w http.ResponseWriter, req *http.Request, actingUserID, token string) {
	vars := mux.Vars(req)
	appID := vars["appid"]
	channelID := vars["channelid"]

	cc := a.conf.GetConfig().SetContextDefaults(&apps.Context{
		ActingUserID: actingUserID,
		ChannelID:    channelID,
	})

	if appID == "" {
		httputils.WriteError(w, errors.New("appid not specified"))
		return
	}

	if err := a.proxy.CacheInvalidateBindings(cc, apps.AppID(appID)); err != nil {
		httputils.WriteError(w, err)
	}
}
