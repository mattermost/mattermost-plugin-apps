package gateway

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (g *gateway) handleWebhook(c *incoming.Request, w http.ResponseWriter, r *http.Request) {
	err := g.doHandleWebhook(c, w, r)
	if err != nil {
		c.Log.WithError(err).Warnw("failed to process remote webhook")
		httputils.WriteError(w, err)
	}
}

func (g *gateway) doHandleWebhook(c *incoming.Request, w http.ResponseWriter, r *http.Request) error {
	appID := appIDVar(r)
	if appID == "" {
		return utils.NewInvalidError("app_id not specified")
	}
	c.SetAppID(appID)

	sreq, err := newHTTPCallRequest(r, c.Config().Get().MaxWebhookSize)
	if err != nil {
		return err
	}
	sreq.Path = mux.Vars(r)["path"]
	c.Log = c.Log.With("call_path", sreq.Path)

	err = g.proxy.NotifyRemoteWebhook(c, appID, *sreq)
	if err != nil {
		return err
	}

	c.Log.Debugf("processed remote webhook")
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
