package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin/handler"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) initSubscriptions(h *handler.Handler) {
	// Subscribe
	h.HandleFunc(path.Subscribe,
		a.Subscribe, h.RequireActingUser, h.RequireFromApp).Methods(http.MethodPost)
	// GetSubscriptions
	h.HandleFunc(path.Subscribe,
		a.GetSubscriptions, h.RequireActingUser, h.RequireFromApp).Methods(http.MethodGet)
	// Unsubscribe
	h.HandleFunc(path.Unsubscribe,
		a.Unsubscribe, h.RequireActingUser, h.RequireFromApp).Methods(http.MethodPost)
}

// Subscribe starts or updates an App subscription to Mattermost events.
//   Path: /api/v1/subscribe
//   Method: POST
//   Input: Subscription
//   Output: None
func (a *restapi) Subscribe(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	a.handleSubscribeCore(r, w, req, true)
}

// GetSubscriptions returns a users current list of subscriptions.
//   Path: /api/v1/subscribe
//   Method: GET
//   Input: None
//   Output: []Subscription
func (a *restapi) GetSubscriptions(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	subs, err := a.appServices.GetSubscriptions(r)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = httputils.WriteJSON(w, subs)
	if err != nil {
		r.Log.WithError(err).Errorf("Error marshaling subscriptions")
	}
}

// Unsubscribe removes an App's subscription to Mattermost events.
//   Path: /api/v1/unsubscribe
//   Method: POST
//   Input: Subscription
//   Output: None
func (a *restapi) Unsubscribe(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	a.handleSubscribeCore(r, w, req, false)
}

func (a *restapi) handleSubscribeCore(r *incoming.Request, w http.ResponseWriter, req *http.Request, isSubscribe bool) {
	status, logMessage, err := func() (int, string, error) {
		var sub apps.Subscription
		if err := json.NewDecoder(req.Body).Decode(&sub); err != nil {
			return http.StatusBadRequest, "Failed to parse Subscription", err
		}
		sub.AppID = r.SourceAppID()
		sub.UserID = r.ActingUserID()

		// TODO replace with an appropriate API-level call that would deduplicate, etc.
		var err error
		if isSubscribe {
			err = a.appServices.Subscribe(r, sub)
		} else {
			err = a.appServices.Unsubscribe(r, sub)
		}

		if err != nil {
			return httputils.ErrorToStatus(err), "Failed to handle subscribe request", err
		}

		return http.StatusOK, "", err
	}()

	if err != nil {
		r.Log.WithError(err).Warnw(logMessage)
		http.Error(w, err.Error(), status)
	}
}
