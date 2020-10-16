package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

const (
	SubscribePath = "/subscribe"
)

type SubscribeResponse struct {
	Error  string            `json:"error,omitempty"`
	Errors map[string]string `json:"errors,omitempty"`
}

type api struct {
	mm   *pluginapi.Client
	apps *apps.Service
}

func Init(router *mux.Router, apps *apps.Service) {
	a := api{
		mm:   apps.Mattermost,
		apps: apps,
	}

	subrouter := router.PathPrefix(constants.APIPath).Subrouter()
	subrouter.HandleFunc("/locations", checkAuthorized(a.handleLocations)).Methods("GET")
	subrouter.HandleFunc(SubscribePath, a.handleSubscribe).Methods("POST", "DELETE")
}

func checkAuthorized(f func(http.ResponseWriter, *http.Request, string)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		actingUserID := req.Header.Get("Mattermost-User-Id")
		if actingUserID == "" {
			httputils.WriteUnauthorizedError(w, errors.New("not authorized"))
			return
		}

		f(w, req, actingUserID)
	}
}

func (a *api) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	var err error
	actingUserID := ""
	// logMessage := ""
	status := http.StatusInternalServerError

	defer func() {
		resp := SubscribeResponse{}
		if err != nil {
			resp.Error = errors.Wrap(err, "failed to subscribe").Error()
			// logMessage = "Error: " + resp.Error
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(resp)
	}()

	actingUserID = r.Header.Get("Mattermost-User-ID")
	if actingUserID == "" {
		err = errors.New("user not logged in")
		status = http.StatusUnauthorized
		return
	}
	// TODO check for sysadmin

	var sub store.Subscription
	if err = json.NewDecoder(r.Body).Decode(&sub); err != nil {
		status = http.StatusUnauthorized
		return
	}

	// TODO replace with an appropriate API-level call that would validate,
	// deduplicate, etc.
	switch r.Method {
	case http.MethodPost:
		err = a.apps.Store.StoreSub(&sub)
	case http.MethodDelete:
		err = a.apps.Store.DeleteSub(&sub)
	default:
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (a *api) handleLocations(w http.ResponseWriter, req *http.Request, actingUserID string) {
	userID := req.URL.Query().Get("user_id")
	if userID == "" {
		httputils.WriteBadRequestError(w, errors.New("no user id"))
		return
	}

	if userID != actingUserID {
		httputils.WriteUnauthorizedError(w, errors.New("user id is not the same"))
		return
	}

	channelID := req.URL.Query().Get("channel_id")
	if channelID == "" {
		httputils.WriteBadRequestError(w, errors.New("no channel id"))
		return
	}

	locations, err := a.apps.API.GetLocations(userID, channelID)
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}

	httputils.WriteJSON(w, locations)
}
