package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) initSubscriptions(h *httpin.Handler) {
	// Subscribe
	h.HandleFunc(path.Subscribe,
		a.Subscribe, httpin.RequireUser, httpin.RequireApp).Methods(http.MethodPost)
	// GetSubscriptions
	h.HandleFunc(path.Subscribe,
		a.GetSubscriptions, httpin.RequireUser, httpin.RequireApp).Methods(http.MethodGet)
	// Unsubscribe
	h.HandleFunc(path.Unsubscribe,
		a.Unsubscribe, httpin.RequireUser, httpin.RequireApp).Methods(http.MethodPost)
}

// Subscribe starts or updates an App subscription to Mattermost events.
//   Path: /api/v1/subscribe
//   Method: POST
//   Input: Subscription
//   Output: None
func (a *restapi) Subscribe(req *incoming.Request, w http.ResponseWriter, r *http.Request) {
	a.handleSubscribeCore(req, w, r, true)
}

// GetSubscriptions returns a users current list of subscriptions.
//   Path: /api/v1/subscribe
//   Method: GET
//   Input: None
//   Output: []Subscription
func (a *restapi) GetSubscriptions(req *incoming.Request, w http.ResponseWriter, r *http.Request) {
	subs, err := a.appServices.GetSubscriptions(req, req.AppID(), req.ActingUserID())
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = httputils.WriteJSON(w, subs)
	if err != nil {
		req.Log.WithError(err).Errorf("Error marshaling subscriptions")
	}
}

// Unsubscribe removes an App's subscription to Mattermost events.
//   Path: /api/v1/unsubscribe
//   Method: POST
//   Input: Subscription
//   Output: None
func (a *restapi) Unsubscribe(req *incoming.Request, w http.ResponseWriter, r *http.Request) {
	a.handleSubscribeCore(req, w, r, false)
}

func (a *restapi) handleSubscribeCore(req *incoming.Request, w http.ResponseWriter, r *http.Request, isSubscribe bool) {
	status, logMessage, err := func() (int, string, error) {
		var sub apps.Subscription
		if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
			return http.StatusBadRequest, "Failed to parse Subscription", err
		}

		sub.AppID = req.AppID()
		sub.UserID = req.ActingUserID()

		// TODO replace with an appropriate API-level call that would deduplicate, etc.
		var err error
		if isSubscribe {
			err = a.appServices.Subscribe(req, sub)
		} else {
			err = a.appServices.Unsubscribe(req, sub)
		}

		if err != nil {
			return httputils.ErrorToStatus(err), "Failed to handle subscribe request", err
		}

		return http.StatusOK, "", err
	}()

	if err != nil {
		req.Log.WithError(err).Warnw(logMessage)
		http.Error(w, err.Error(), status)
	}
}
