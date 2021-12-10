package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) initOAuth2Store(api *mux.Router) {
	// TODO appid should come from OAuth2 user session, see
	// https://mattermost.atlassian.net/browse/MM-34377
	api.HandleFunc(path.OAuth2App+"/{appid}",
		proxy.RequireUser(a.OAuth2StoreApp)).Methods("PUT", "POST")
	api.HandleFunc(path.OAuth2User+"/{appid}",
		proxy.RequireUser(a.OAuth2StoreUser)).Methods("PUT", "POST")
	api.HandleFunc(path.OAuth2User+"/{appid}",
		proxy.RequireUser(a.OAuth2GetUser)).Methods("GET")
}

func (a *restapi) OAuth2StoreApp(w http.ResponseWriter, r *http.Request, in proxy.Incoming) {
	oapp := apps.OAuth2App{}
	err := json.NewDecoder(r.Body).Decode(&oapp)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	err = a.appServices.StoreOAuth2App(appIDVar(r), in.ActingUserID, oapp)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) OAuth2StoreUser(w http.ResponseWriter, r *http.Request, in proxy.Incoming) {
	data, err := httputils.LimitReadAll(r.Body, MaxKVStoreValueLength)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	err = a.appServices.StoreOAuth2User(appIDVar(r), in.ActingUserID, data)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) OAuth2GetUser(w http.ResponseWriter, r *http.Request, in proxy.Incoming) {
	var v interface{}
	err := a.appServices.GetOAuth2User(appIDVar(r), in.ActingUserID, &v)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	_ = httputils.WriteJSON(w, v)
}
