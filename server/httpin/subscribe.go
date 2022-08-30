package httpin

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

// GetSubscriptions returns a users current list of subscriptions.
//   Path: /api/v1/subscribe
//   Method: GET
//   Input: None
//   Output: []Subscription
func (s *Service) GetSubscriptions(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	subs, err := s.AppServices.GetSubscriptions(r)
	if err != nil {
		http.Error(w, err.Error(), httputils.ErrorToStatus(err))
		return
	}
	_ = httputils.WriteJSON(w, subs)
}

// Subscribe starts or updates an App subscription to Mattermost events.
//   Path: /api/v1/subscribe
//   Method: POST
//   Input: Subscription
//   Output: None
func (s *Service) Subscribe(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	var sub apps.Subscription
	if err := json.NewDecoder(req.Body).Decode(&sub); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.AppServices.Subscribe(r, sub); err != nil {
		http.Error(w, err.Error(), httputils.ErrorToStatus(err))
		return
	}
}

// Unsubscribe removes an App's subscription to Mattermost events.
//   Path: /api/v1/unsubscribe
//   Method: POST
//   Input: Subscription
//   Output: None
func (s *Service) Unsubscribe(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	var e apps.Event
	if err := json.NewDecoder(req.Body).Decode(&e); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.AppServices.Unsubscribe(r, e); err != nil {
		http.Error(w, err.Error(), httputils.ErrorToStatus(err))
		return
	}
}
