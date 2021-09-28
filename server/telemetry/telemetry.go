package telemetry

import (
	mmtelemetry "github.com/mattermost/mattermost-plugin-api/experimental/telemetry"
)

type Telemetry struct {
	tracker mmtelemetry.Tracker
}

func NewTelemetry(tracker mmtelemetry.Tracker) *Telemetry {
	return &Telemetry{
		tracker: tracker,
	}
}

func (t *Telemetry) UpdateTracker(tracker mmtelemetry.Tracker) {
	t.tracker = tracker
}

func (t *Telemetry) TrackInstall(appID, appType string) {
	if t == nil || t.tracker == nil {
		return
	}

	_ = t.tracker.TrackEvent("install", map[string]interface{}{
		"appID":   appID,
		"appType": appType,
	})
}

func (t *Telemetry) TrackUninstall(appID, appType string) {
	if t == nil || t.tracker == nil {
		return
	}

	_ = t.tracker.TrackEvent("uninstall", map[string]interface{}{
		"appID":   appID,
		"appType": appType,
	})
}

func (t *Telemetry) TrackCall(appID, location, actingUserID, callType string) {
	if t == nil || t.tracker == nil {
		return
	}

	_ = t.tracker.TrackUserEvent("call", actingUserID, map[string]interface{}{
		"appID":    appID,
		"location": location,
		"type":     callType,
	})
}

func (t *Telemetry) TrackOAuthComplete(appID, actingUserID string) {
	if t == nil || t.tracker == nil {
		return
	}

	_ = t.tracker.TrackUserEvent("oauthComplete", actingUserID, map[string]interface{}{
		"appID": appID,
	})
}
