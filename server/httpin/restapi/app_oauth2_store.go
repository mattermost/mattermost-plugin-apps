package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) initOAuth2Store(h *httpin.Handler) {
	h.HandleFunc(path.OAuth2App,
		a.OAuth2StoreApp, httpin.RequireSysadmin, httpin.RequireApp).Methods(http.MethodPut, http.MethodPost)
	h.HandleFunc(path.OAuth2User,
		a.OAuth2StoreUser, httpin.RequireUser, httpin.RequireApp).Methods(http.MethodPut, http.MethodPost)
	h.HandleFunc(path.OAuth2User,
		a.OAuth2GetUser, httpin.RequireUser, httpin.RequireApp).Methods(http.MethodGet)
}

func (a *restapi) OAuth2StoreApp(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	data, err := httputils.LimitReadAll(req.Body, MaxKVStoreValueLength)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	err = a.appServices.StoreOAuth2App(r, r.AppID(), r.ActingUserID(), data)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) OAuth2StoreUser(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	data, err := httputils.LimitReadAll(req.Body, MaxKVStoreValueLength)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	err = a.appServices.StoreOAuth2User(r, r.AppID(), r.ActingUserID(), data)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) OAuth2GetUser(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	data, err := a.appServices.GetOAuth2User(r, r.AppID(), r.ActingUserID())
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	_, _ = w.Write(data)
}
