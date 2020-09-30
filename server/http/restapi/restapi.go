package restapi

import (
	"encoding/json"
	"net/http"

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

func (a *api) handleSubscribe(w http.ResponseWriter, req *http.Request) {
	var err error

	// actingUserID := req.Header.Get("Mattermost-User-Id")
	// fmt.Printf("actingUserID = %+v\n", actingUserID)
	// if actingUserID == "" {
	// 	// err = errors.New("user not logged in")
	// 	status = http.StatusUnauthorized
	// 	return
	// }

	var subRequest apps.Subscription
	if err = json.NewDecoder(req.Body).Decode(&subRequest); err != nil {
		// return respondErr(w, http.StatusInternalServerError, err)
		return
	}
	subs := apps.NewSubscriptions(a.mm, a.configurator)

	switch req.Method {
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
