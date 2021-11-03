package gateway

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (g *gateway) handleWebhook(w http.ResponseWriter, req *http.Request) {
	log, err := g.doHandleWebhook(w, req, g.conf.Logger())
	if err != nil {
		log.WithError(err).Warnw("failed to process remote webhook")
		httputils.WriteError(w, err)
	}
}

func (g *gateway) doHandleWebhook(w http.ResponseWriter, req *http.Request, log utils.Logger) (utils.Logger, error) {
	appID := appIDVar(req)
	if appID == "" {
		return log, utils.NewInvalidError("app_id not specified")
	}
	log = log.With("app_id", appID)

	sreq, err := newHTTPCallRequest(req, g.conf.Get().MaxWebhookSize)
	if err != nil {
		return log, err
	}
	sreq.Path = mux.Vars(req)["path"]
	log = log.With("path", sreq.Path)

	err = g.proxy.NotifyRemoteWebhook(appID, *sreq)
	if err != nil {
		return log, err
	}

	log.Debugf("processed remote webhook")
	return log, nil
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
