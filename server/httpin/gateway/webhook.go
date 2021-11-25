package gateway

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (g *gateway) handleWebhook(req *incoming.Request, w http.ResponseWriter, r *http.Request) {
	err := g.doHandleWebhook(req, w, r)
	if err != nil {
		req.Log.WithError(err).Warnw("failed to process remote webhook")
		httputils.WriteError(w, err)
	}
}

func (g *gateway) doHandleWebhook(req *incoming.Request, _ http.ResponseWriter, r *http.Request) error {
	appID := appIDVar(r)
	if appID == "" {
		return utils.NewInvalidError("app_id not specified")
	}
	req.SetAppID(appID)

	sreq, err := newHTTPCallRequest(r, g.conf.Get().MaxWebhookSize)
	if err != nil {
		return err
	}
	sreq.Path = mux.Vars(r)["path"]
	req.Log = req.Log.With("call_path", sreq.Path)

	err = g.proxy.NotifyRemoteWebhook(req, appID, *sreq)
	if err != nil {
		return err
	}

	req.Log.Debugf("processed remote webhook")
	return nil
}

func newHTTPCallRequest(r *http.Request, limit int64) (*apps.HTTPCallRequest, error) {
	data, err := httputils.LimitReadAll(r.Body, limit)
	if err != nil {
		return nil, err
	}

	sreq := apps.HTTPCallRequest{
		HTTPMethod: r.Method,
		Path:       r.URL.Path,
		RawQuery:   r.URL.RawQuery,
		Body:       string(data),
		Headers:    map[string]string{},
	}
	for key := range r.Header {
		sreq.Headers[key] = r.Header.Get(key)
	}

	return &sreq, nil
}
