package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy/request"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) initOAuth2Store(api *mux.Router, c *request.Context) {
	// TODO appid should come from OAuth2 user session, see
	// https://mattermost.atlassian.net/browse/MM-34377
	api.Handle(path.OAuth2App+"/{appid}",
		request.AddContext(a.OAuth2StoreApp, c).RequireSysadmin()).Methods(http.MethodPut, http.MethodPost)
	api.Handle(path.OAuth2User+"/{appid}",
		request.AddContext(a.OAuth2StoreUser, c).RequireUser()).Methods(http.MethodPut, http.MethodPost)
	api.Handle(path.OAuth2User+"/{appid}",
		request.AddContext(a.OAuth2GetUser, c).RequireUser()).Methods(http.MethodGet)
}

func (a *restapi) OAuth2StoreApp(c *request.Context, w http.ResponseWriter, r *http.Request) {
	oapp := apps.OAuth2App{}
	err := json.NewDecoder(r.Body).Decode(&oapp)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	err = a.appServices.StoreOAuth2App(appIDVar(r), c.ActingUserID(), oapp)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) OAuth2StoreUser(c *request.Context, w http.ResponseWriter, r *http.Request) {
	data, err := httputils.LimitReadAll(r.Body, MaxKVStoreValueLength)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	err = a.appServices.StoreOAuth2User(appIDVar(r), c.ActingUserID(), data)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) OAuth2GetUser(c *request.Context, w http.ResponseWriter, r *http.Request) {
	var v interface{}
	err := a.appServices.GetOAuth2User(appIDVar(r), c.ActingUserID(), &v)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	_ = httputils.WriteJSON(w, v)
}
