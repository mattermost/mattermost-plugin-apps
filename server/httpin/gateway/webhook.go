package gateway

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (g *gateway) handleWebhook(w http.ResponseWriter, r *http.Request) {
	log, err := g.doHandleWebhook(w, r, g.conf.Logger())
	if err != nil {
		log.WithError(err).Warnw("failed to process remote webhook")
		httputils.WriteError(w, err)
	}
}

func (g *gateway) doHandleWebhook(w http.ResponseWriter, r *http.Request, log utils.Logger) (utils.Logger, error) {
	appID := appIDVar(r)
	if appID == "" {
		return log, utils.NewInvalidError("app_id not specified")
	}
	log = log.With("app_id", appID)

	sreq, err := newHTTPCallRequest(r, g.conf.Get().MaxWebhookSize)
	if err != nil {
		return log, err
	}
	sreq.Path = mux.Vars(r)["path"]
	log = log.With("path", sreq.Path)

	err = g.proxy.NotifyRemoteWebhook(appID, *sreq)
	if err != nil {
		return log, err
	}

	log.Debugf("processed remote webhook")
	return log, nil
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
