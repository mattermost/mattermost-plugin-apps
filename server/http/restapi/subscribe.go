package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (a *restapi) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	a.handleSubscribeCore(w, r, true)
}

func (a *restapi) handleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	a.handleSubscribeCore(w, r, false)
}

func (a *restapi) handleSubscribeCore(w http.ResponseWriter, r *http.Request, isSubscribe bool) {
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

	actingUserID = r.Header.Get("Mattermost-User-ID")

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
		err = a.appServices.Subscribe(actingUserID, &sub)
	} else {
		err = a.appServices.Unsubscribe(actingUserID, &sub)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
