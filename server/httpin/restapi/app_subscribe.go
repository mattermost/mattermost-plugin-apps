package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) initSubscriptions(api *mux.Router, c *incoming.Request) {
	// Subscribe
	api.Handle(path.Subscribe,
		incoming.AddContext(a.Subscribe, c).RequireUser().RequireApp()).Methods(http.MethodPost)
	// GetSubscriptions
	api.Handle(path.Subscribe,
		incoming.AddContext(a.GetSubscriptions, c).RequireUser().RequireApp()).Methods(http.MethodGet)
	// Unsubscribe
	api.Handle(path.Unsubscribe,
		incoming.AddContext(a.Unsubscribe, c).RequireUser().RequireApp()).Methods(http.MethodPost)
}

// Subscribe starts or updates an App subscription to Mattermost events.
//   Path: /api/v1/subscribe
//   Method: POST
//   Input: Subscription
//   Output: None
func (a *restapi) Subscribe(c *incoming.Request, w http.ResponseWriter, r *http.Request) {
	a.handleSubscribeCore(c, w, r, true)
}

// GetSubscriptions returns a users current list of subscriptions.
//   Path: /api/v1/subscribe
//   Method: GET
//   Input: None
//   Output: []Subscription
func (a *restapi) GetSubscriptions(c *incoming.Request, w http.ResponseWriter, r *http.Request) {
	subs, err := a.appServices.GetSubscriptions(c.AppID(), c.ActingUserID())
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = httputils.WriteJSON(w, subs)
	if err != nil {
		c.Log.WithError(err).Errorf("Error marshaling subscriptions")
	}
}

// Unsubscribe removes an App's subscription to Mattermost events.
//   Path: /api/v1/unsubscribe
//   Method: POST
//   Input: Subscription
//   Output: None
func (a *restapi) Unsubscribe(c *incoming.Request, w http.ResponseWriter, r *http.Request) {
	a.handleSubscribeCore(c, w, r, false)
}

func (a *restapi) handleSubscribeCore(c *incoming.Request, w http.ResponseWriter, r *http.Request, isSubscribe bool) {
	status, logMessage, err := func() (int, string, error) {
		var sub apps.Subscription
		if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
			return http.StatusBadRequest, "Failed to parse Subscription", err
		}

		sub.AppID = c.AppID()
		sub.UserID = c.ActingUserID()

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
		c.Log.WithError(err).Warnw(logMessage)
		http.Error(w, err.Error(), status)
	}
}
