package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) initOAuth2Store(rh *httpin.Handler) {
	rh.HandleFunc(path.OAuth2App,
		a.OAuth2StoreApp, httpin.RequireSysadmin, httpin.RequireApp).Methods(http.MethodPut, http.MethodPost)
	rh.HandleFunc(path.OAuth2User,
		a.OAuth2StoreUser, httpin.RequireUser, httpin.RequireApp).Methods(http.MethodPut, http.MethodPost)
	rh.HandleFunc(path.OAuth2User,
		a.OAuth2GetUser, httpin.RequireUser, httpin.RequireApp).Methods(http.MethodGet)
}

func (a *restapi) OAuth2StoreApp(req *incoming.Request, w http.ResponseWriter, r *http.Request) {
	oapp := apps.OAuth2App{}
	err := json.NewDecoder(r.Body).Decode(&oapp)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	err = a.appServices.StoreOAuth2App(req.AppID(), req.ActingUserID(), oapp)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) OAuth2StoreUser(req *incoming.Request, w http.ResponseWriter, r *http.Request) {
	data, err := httputils.LimitReadAll(r.Body, MaxKVStoreValueLength)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	err = a.appServices.StoreOAuth2User(req.AppID(), req.ActingUserID(), data)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) OAuth2GetUser(req *incoming.Request, w http.ResponseWriter, r *http.Request) {
	var v interface{}
	err := a.appServices.GetOAuth2User(req.AppID(), req.ActingUserID(), &v)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	_ = httputils.WriteJSON(w, v)
}
