package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"

	"github.com/gorilla/mux"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
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
	// subs         *apps.Subscriptions
	configurator configurator.Service
}

func Init(router *mux.Router, apps *apps.Service) {
	a := api{
		mm:   apps.Mattermost,
		apps: apps,
	}

	subrouter := router.PathPrefix(constants.APIPath).Subrouter()
	subrouter.HandleFunc(SubscribePath, a.handleSubscribe).Methods("POST", "DELETE")
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
	// <><> TODO check for sysadmin

	var subRequest apps.Subscription
	if err = json.NewDecoder(r.Body).Decode(&subRequest); err != nil {
		status = http.StatusUnauthorized
		return
	}
	subs := apps.NewSubscriptions(a.mm, a.configurator)

	switch r.Method {
	case http.MethodPost:
		err = subs.StoreSub(subRequest)
	case http.MethodDelete:
		err = subs.DeleteSub(subRequest)
	default:
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
