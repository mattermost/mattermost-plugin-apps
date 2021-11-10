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
	api.Handle(path.OAuth2App,
		request.AddContext(a.OAuth2StoreApp, c).RequireSysadmin().RequireApp()).Methods(http.MethodPut, http.MethodPost)
	api.Handle(path.OAuth2User,
		request.AddContext(a.OAuth2StoreUser, c).RequireUser().RequireApp()).Methods(http.MethodPut, http.MethodPost)
	api.Handle(path.OAuth2User,
		request.AddContext(a.OAuth2GetUser, c).RequireUser().RequireApp()).Methods(http.MethodGet)
}

func (a *restapi) OAuth2StoreApp(c *request.Context, w http.ResponseWriter, r *http.Request) {
	oapp := apps.OAuth2App{}
	err := json.NewDecoder(r.Body).Decode(&oapp)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	err = a.appServices.StoreOAuth2App(c.AppID(), c.ActingUserID(), oapp)
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
	err = a.appServices.StoreOAuth2User(c.AppID(), c.ActingUserID(), data)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) OAuth2GetUser(c *request.Context, w http.ResponseWriter, r *http.Request) {
	var v interface{}
	err := a.appServices.GetOAuth2User(c.AppID(), c.ActingUserID(), &v)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	_ = httputils.WriteJSON(w, v)
}
