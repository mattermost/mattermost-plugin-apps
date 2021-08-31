package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) oauth2StoreApp(w http.ResponseWriter, r *http.Request) {
	oapp := apps.OAuth2App{}
	err := json.NewDecoder(r.Body).Decode(&oapp)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	err = a.appServices.StoreOAuth2App(appIDVar(r), actingID(r), oapp)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) oauth2StoreUser(w http.ResponseWriter, r *http.Request) {
	data, err := httputils.LimitReadAll(r.Body, MaxKVStoreValueLength)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	err = a.appServices.StoreOAuth2User(appIDVar(r), actingID(r), data)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) oauth2GetUser(w http.ResponseWriter, r *http.Request) {
	var v interface{}
	err := a.appServices.GetOAuth2User(appIDVar(r), actingID(r), &v)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	httputils.WriteJSON(w, v)
}
