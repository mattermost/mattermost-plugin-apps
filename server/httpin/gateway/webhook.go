package gateway

import (
	"net/http"

	"github.com/gorilla/mux"

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

		sreq, err := httputils.ServerlessRequestFromHTTP(req, g.conf.Get().MaxWebhookSize)
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
