package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) initSubscriptions(api *mux.Router, mm *pluginapi.Client) {
	// Subscribe
	api.HandleFunc(path.Subscribe,
		proxy.RequireUser(a.Subscribe)).Methods(http.MethodPost)
	// GetSubscriptions
	api.HandleFunc(path.Subscribe,
		proxy.RequireUser(a.GetSubscriptions)).Methods("GET")
	// Unsubscribe
	api.HandleFunc(path.Unsubscribe,
		proxy.RequireUser(a.Unsubscribe)).Methods(http.MethodPost)
}

// Subscribe starts or updates an App subscription to Mattermost events.
//   Path: /api/v1/subscribe
//   Method: POST
//   Input: Subscription
//   Output: None
func (a *restapi) Subscribe(w http.ResponseWriter, r *http.Request, in proxy.Incoming) {
	a.handleSubscribeCore(w, r, in, true)
}

// GetSubscriptions returns a users current list of subscriptions.
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

	err = httputils.WriteJSON(w, subs)
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
	log := a.conf.Logger()
	status, logMessage, err := func() (int, string, error) {
		var sub apps.Subscription
		if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
			return http.StatusBadRequest, "Failed to parse Subscription", err
		}

		sub.UserID = in.ActingUserID

		if err := sub.Validate(); err != nil {
			return http.StatusBadRequest, "Invalid Subscription", err
		}

		// TODO replace with an appropriate API-level call that would validate,
		// deduplicate, etc.
		var err error
		if isSubscribe {
			err = a.appServices.Subscribe(sub)
		} else {
			err = a.appServices.Unsubscribe(sub)
		}

		if err != nil {
			return httputils.ErrorToStatus(err), "Failed to handle subscribe request", err
		}

		return http.StatusOK, "", err
	}()

	if err != nil {
		log.WithError(err).Warnw(logMessage)
		http.Error(w, err.Error(), status)
	}
}
