package gateway

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (g *gateway) remoteOAuth2Connect(w http.ResponseWriter, req *http.Request, in proxy.Incoming) {
	appID := appIDVar(req)

	if appID == "" {
		httputils.WriteError(w, utils.NewInvalidError("app_id not specified"))
		return
	}

	connectURL, err := g.proxy.GetRemoteOAuth2ConnectURL(in, appID)
	if err != nil {
		g.log.WithError(err).Warnw("Failed to get remote OuAuth2 connect URL",
			"app_id", appID,
			"acting_user_id", in.ActingUserID)
		httputils.WriteError(w, err)
		return
	}

	http.Redirect(w, req, connectURL, http.StatusTemporaryRedirect)
}

func (g *gateway) remoteOAuth2Complete(w http.ResponseWriter, req *http.Request, in proxy.Incoming) {
	appID := appIDVar(req)

	if appID == "" {
		httputils.WriteError(w, utils.NewInvalidError("app_id not specified"))
		return
	}

	q := req.URL.Query()
	urlValues := map[string]interface{}{}
	for key := range q {
		urlValues[key] = q.Get(key)
	}

	err := g.proxy.CompleteRemoteOAuth2(in, appID, urlValues)
	if err != nil {
		g.log.WithError(err).Warnw("Failed to complete remote OuAuth2",
			"app_id", appID,
			"acting_user_id", in.ActingUserID)
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
			<p>Completed connecting your account. Please close this window.</p>
		</body>
	</html>
	`))
}
