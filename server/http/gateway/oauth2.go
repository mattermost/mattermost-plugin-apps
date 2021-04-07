package gateway

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (g *gateway) remoteOAuth2Connect(w http.ResponseWriter, req *http.Request, sessionID, actingUserID string) {
	if appIDVar(req) == "" {
		httputils.WriteError(w, utils.NewInvalidError("app_id not specified"))
		return
	}

	connectURL, err := g.proxy.GetRemoteOAuth2ConnectURL(sessionID, actingUserID, appIDVar(req))
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	http.Redirect(w, req, connectURL, http.StatusTemporaryRedirect)
}

func (g *gateway) remoteOAuth2Complete(w http.ResponseWriter, req *http.Request, sessionID, actingUserID string) {
	if appIDVar(req) == "" {
		httputils.WriteError(w, utils.NewInvalidError("app_id not specified"))
		return
	}

	q := req.URL.Query()
	urlValues := map[string]interface{}{}
	for key := range q {
		urlValues[key] = q.Get(key)
	}

	err := g.proxy.CompleteRemoteOAuth2(sessionID, actingUserID, appIDVar(req), urlValues)
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
