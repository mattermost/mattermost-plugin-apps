package restapi

import (
	// nolint:gosec

	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func(a *restapi) getClient(userId string, w http.ResponseWriter, r *http.Request) *mmclient.Client {
	var err error

	sessionID := r.Header.Get("MM_SESSION_ID")
	if sessionID == "" {
		err = errors.New("no user session")
		httputils.WriteUnauthorizedError(w, err)
		return nil
	}

	session, err := a.api.Mattermost.Session.Get(sessionID)
	if err != nil {
		httputils.WriteUnauthorizedError(w, err)
		return nil
	}

	conf := a.api.Configurator.GetConfig()
	token := string(apps.SessionToken(session.Token))
	client := mmclient.NewClient(userId, token, conf.MattermostSiteURL)

	return client
}