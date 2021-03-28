package restapi

import (
	// nolint:gosec

	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) oauth2CreateState(w http.ResponseWriter, r *http.Request) {
	state, err := a.appServices.CreateOAuth2State(actingID(r))
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	httputils.WriteJSON(w, state)
}

func (a *restapi) oauth2ValidateState(w http.ResponseWriter, r *http.Request) {
	state := ""
	err := json.NewDecoder(r.Body).Decode(&state)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	err = a.appServices.ValidateOAuth2State(actingID(r), state)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	httputils.WriteJSON(w, true)
}

func (a *restapi) oauth2StoreApp(w http.ResponseWriter, r *http.Request) {
	oapp := apps.OAuth2App{}
	err := json.NewDecoder(r.Body).Decode(&oapp)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	err = a.appServices.StoreOAuth2App(actingID(r), oapp)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) oauth2StoreUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["appid"]
	data, err := io.ReadAll(r.Body)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	err = a.appServices.StoreOAuth2User(apps.AppID(id), actingID(r), data)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) oauth2GetUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["appid"]
	v := map[string]interface{}{}
	err := a.appServices.GetOAuth2User(apps.AppID(id), actingID(r), &v)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	httputils.WriteJSON(w, v)
}
