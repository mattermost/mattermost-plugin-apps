package gateway

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (g *gateway) remoteOAuth2Connect(c *incoming.Request, w http.ResponseWriter, r *http.Request) {
	appID := appIDVar(r)
	if appID == "" {
		httputils.WriteError(w, utils.NewInvalidError("app_id not specified"))
		return
	}
	c.SetAppID(appID)

	connectURL, err := g.proxy.GetRemoteOAuth2ConnectURL(c, appID)
	if err != nil {
		c.Log.WithError(err).Warnf("Failed to get remote OAuth2 connect URL")
		httputils.WriteError(w, err)
		return
	}

	http.Redirect(w, r, connectURL, http.StatusTemporaryRedirect)
}

func (g *gateway) remoteOAuth2Complete(c *incoming.Request, w http.ResponseWriter, r *http.Request) {
	appID := appIDVar(r)
	if appID == "" {
		httputils.WriteError(w, utils.NewInvalidError("app_id not specified"))
		return
	}
	c.SetAppID(appID)

	q := r.URL.Query()
	urlValues := map[string]interface{}{}
	for key := range q {
		urlValues[key] = q.Get(key)
	}

	err := g.proxy.CompleteRemoteOAuth2(c, appID, urlValues)
	if err != nil {
		c.Log.WithError(err).Warnf("Failed to complete remote OAuth2")
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
