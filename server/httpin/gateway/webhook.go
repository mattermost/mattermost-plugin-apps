package gateway

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (g *gateway) handleWebhook(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	err := g.doHandleWebhook(r, w, req)
	if err != nil {
		r.Log.WithError(err).Warnw("failed to process remote webhook")
		httputils.WriteError(w, err)
	}
}

func (g *gateway) doHandleWebhook(r *incoming.Request, _ http.ResponseWriter, req *http.Request) error {
	appID := appIDVar(req)
	if appID == "" {
		return utils.NewInvalidError("app_id not specified")
	}

	sreq, err := newHTTPCallRequest(req, g.Config.Get().MaxWebhookSize)
	if err != nil {
		return err
	}
	sreq.Path = mux.Vars(req)["path"]
	r.Log = r.Log.With("call_path", sreq.Path)

	err = g.Proxy.NotifyRemoteWebhook(r, appID, *sreq)
	if err != nil {
		return err
	}

	r.Log.Debugf("processed remote webhook")
	return nil
}

func newHTTPCallRequest(req *http.Request, limit int64) (*apps.HTTPCallRequest, error) {
	data, err := httputils.LimitReadAll(req.Body, limit)
	if err != nil {
		return nil, err
	}

	sreq := apps.HTTPCallRequest{
		HTTPMethod: req.Method,
		Path:       req.URL.Path,
		RawQuery:   req.URL.RawQuery,
		Body:       string(data),
		Headers:    map[string]string{},
	}
	for key := range req.Header {
		sreq.Headers[key] = req.Header.Get(key)
	}

	return &sreq, nil
}
