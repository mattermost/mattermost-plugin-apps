package gateway

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (g *gateway) handleWebhook(w http.ResponseWriter, req *http.Request) {
	log := g.conf.Logger()
	err := func() error {
		appID := appIDVar(req)
		if appID == "" {
			return utils.NewInvalidError("app_id not specified")
		}
		log = log.With("app_id", appID)

		sreq, err := serverlessRequestFromHTTP(req, g.conf.Get().MaxWebhookSize)
		if err != nil {
			return err
		}
		sreq.Path = mux.Vars(req)["path"]
		log = log.With("path", sreq.Path)

		err = g.proxy.NotifyRemoteWebhook(appID, *sreq)
		if err != nil {
			return err
		}

		log.Debugf("processed remote webhook")
		return nil
	}()

	if err != nil {
		log.WithError(err).Warnw("failed to process remote webhook")
		httputils.WriteError(w, err)
	}
}

func serverlessRequestFromHTTP(req *http.Request, limit int64) (*apps.ServerlessRequest, error) {
	data, err := httputils.LimitReadAll(req.Body, limit)
	if err != nil {
		return nil, err
	}

	sreq := apps.ServerlessRequest{
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
