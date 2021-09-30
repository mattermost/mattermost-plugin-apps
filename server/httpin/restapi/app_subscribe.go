package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

func (a *restapi) initSubscriptions(api *mux.Router, mm *pluginapi.Client) {
	// Subscribe
	api.HandleFunc(path.Subscribe,
		proxy.RequireSysadmin(mm, a.Subscribe)).Methods("POST")
	// GetSubscriptions
	api.HandleFunc(path.Subscribe,
		proxy.RequireSysadmin(mm, a.GetSubscriptions)).Methods("GET")
	// Unsubscribe
	api.HandleFunc(path.Unsubscribe,
		proxy.RequireSysadmin(mm, a.Unsubscribe)).Methods("POST")
}

// Subscribe starts or updates an App subscription to Mattermost events.
//   Path: /api/v1/subscribe
//   Method: POST
//   Input: Subscription
//   Output: None
func (a *restapi) Subscribe(w http.ResponseWriter, r *http.Request, in proxy.Incoming) {
	a.handleSubscribeCore(w, r, in, true)
}

// GetSubscriptions returns the App's current list of subscriptions.
//   Path: /api/v1/subscribe
//   Method: GET
//   Input: None
//   Output: []Subscription
func (a *restapi) GetSubscriptions(w http.ResponseWriter, r *http.Request, in proxy.Incoming) {
	subs, err := a.appServices.GetSubscriptions(in.ActingUserID)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(subs)
	if err != nil {
		a.conf.Logger().WithError(err).Errorf("Error marshaling subscriptions")
	}
}

// Unsubscribe removes an App's subscription to Mattermost events.
//   Path: /api/v1/unsubscribe
//   Method: POST
//   Input: Subscription
//   Output: None
func (a *restapi) Unsubscribe(w http.ResponseWriter, r *http.Request, in proxy.Incoming) {
	a.handleSubscribeCore(w, r, in, false)
}

func (a *restapi) handleSubscribeCore(w http.ResponseWriter, r *http.Request, in proxy.Incoming, isSubscribe bool) {
	var err error
	actingUserID := ""
	// logMessage := ""
	status := http.StatusOK

	defer func() {
		resp := apps.SubscriptionResponse{}
		if err != nil {
			resp.Error = errors.Wrap(err, "failed operation").Error()
			status = http.StatusInternalServerError
			// logMessage = "Error: " + resp.Error
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(resp.ToJSON())
	}()

	actingUserID = in.ActingUserID
	if actingUserID == "" {
		err = errors.New("user not logged in")
		status = http.StatusUnauthorized
		return
	}

	var sub apps.Subscription
	if err = json.NewDecoder(r.Body).Decode(&sub); err != nil {
		status = http.StatusUnauthorized
		return
	}

	// TODO replace with an appropriate API-level call that would validate,
	// deduplicate, etc.
	if isSubscribe {
		err = a.appServices.Subscribe(actingUserID, sub)
	} else {
		err = a.appServices.Unsubscribe(actingUserID, sub)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
