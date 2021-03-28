package gateway

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (g *gateway) remoteOAuth2Redirect(w http.ResponseWriter, req *http.Request, sessionID, actingUserID string) {
	vars := mux.Vars(req)

	appID := vars["app_id"]
	if appID == "" {
		httputils.WriteError(w, utils.NewInvalidError("app_id not specified"))
		return
	}

	redirectURL, err := g.proxy.GetRemoteOAuth2RedirectURL(sessionID, actingUserID, apps.AppID(appID))
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	http.Redirect(w, req, redirectURL, http.StatusTemporaryRedirect)
}

func (g *gateway) remoteOAuth2Complete(w http.ResponseWriter, req *http.Request, sessionID, actingUserID string) {
	vars := mux.Vars(req)

	appID := vars["app_id"]
	if appID == "" {
		httputils.WriteError(w, utils.NewInvalidError("app_id not specified"))
		return
	}

	q := req.URL.Query()
	urlValues := map[string]interface{}{}
	for key := range q {
		urlValues[key] = q.Get(key)
	}

	err := g.proxy.CompleteRemoteOAuth2(sessionID, actingUserID, apps.AppID(appID), urlValues)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(`
	<!DOCTYPE html>
	<html>
		<head>
			<script>
				window.close();
			</script>
		</head>
		<body>
			<p>Completed connecting to Google. Please close this window.</p>
		</body>
	</html>
	`))
}
