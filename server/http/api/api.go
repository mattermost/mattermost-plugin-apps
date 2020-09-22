package api

import (
	"encoding/json"
	"fmt"
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
		mm:           apps.Mattermost,
		configurator: apps.Config,
	}

	subrouter := router.PathPrefix(constants.APIPath + SubscribePath).Subrouter()
	subrouter.HandleFunc("/channel_created", a.handleSubscribeChannelCreated).Methods("POST")
	subrouter.HandleFunc("/post_created", a.handleSubscribePostCreated).Methods("POST")
	subrouter.HandleFunc("/user_created", a.handleSubscribeUserCreated).Methods("POST")
	subrouter.HandleFunc("/user_updated", a.handleSubscribeUserUpdated).Methods("POST")
	subrouter.HandleFunc("/user_joined_channel", a.handleSubscribeUserJoinedChannel).Methods("POST")
	subrouter.HandleFunc("/user_left_channel", a.handleSubscribeUserLeftChannel).Methods("POST")
	subrouter.HandleFunc("/user_joined_team", a.handleSubscribeUserJoinedTeam).Methods("POST")
	subrouter.HandleFunc("/user_left_team", a.handleSubscribeUserLeftTeam).Methods("POST")
}

func (a *api) handleSubscribeUserJoinedChannel(w http.ResponseWriter, req *http.Request) {
	var err error

	//create a dummy
	var subRequest apps.Subscription
	subRequest.Subject = "user_joined_channel"
	subRequest.AppID = "AppID"
	subRequest.SubscriptionID = "SubsID"
	subRequest.ChannelID = "some_channel_idXXX2"
	subRequestD, _ := json.MarshalIndent(subRequest, "", "    ")
	fmt.Printf("subRequest = %+v\n", string(subRequestD))

	subs := apps.NewSubscriptions(a.mm, a.configurator)
	err = subs.StoreSubscription(subRequest.Subject, subRequest, subRequest.ChannelID)
	if err != nil {
		// status = http.StatusBadRequest
		return
	}

	// actingUserID := req.Header.Get("Mattermost-User-Id")
	// fmt.Printf("actingUserID = %+v\n", actingUserID)
	// if actingUserID == "" {
	// 	// err = errors.New("user not logged in")
	// 	status = http.StatusUnauthorized
	// 	return
	// }

	// var subRequest apps.Subscription
	// err = json.NewDecoder(req.Body).Decode(&subRequest)
	// if err != nil {
	// 	status = http.StatusBadRequest
	// 	return
	// }
}

func (a *api) handleSubscribeUserLeftChannel(w http.ResponseWriter, req *http.Request) {
}

func (a *api) handleSubscribeChannelCreated(w http.ResponseWriter, req *http.Request) {
}

func (a *api) handleSubscribePostCreated(w http.ResponseWriter, req *http.Request) {
}

func (a *api) handleSubscribeUserCreated(w http.ResponseWriter, req *http.Request) {
}

func (a *api) handleSubscribeUserJoinedTeam(w http.ResponseWriter, req *http.Request) {
}

func (a *api) handleSubscribeUserLeftTeam(w http.ResponseWriter, req *http.Request) {
}

func (a *api) handleSubscribeUserUpdated(w http.ResponseWriter, req *http.Request) {
}
