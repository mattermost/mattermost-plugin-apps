package gateway

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (g *gateway) remoteOAuth2Connect(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	connectURL, err := g.Proxy.InvokeGetRemoteOAuth2ConnectURL(r)
	if err != nil {
		r.Log.WithError(err).Warnf("Failed to get remote OAuth2 connect URL")
		httputils.WriteError(w, err)
		return
	}
	http.Redirect(w, req, connectURL, http.StatusTemporaryRedirect)
}

func (g *gateway) remoteOAuth2Complete(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	urlValues := map[string]interface{}{}
	for key := range q {
		urlValues[key] = q.Get(key)
	}

	err := g.Proxy.InvokeCompleteRemoteOAuth2(r, urlValues)
	if err != nil {
		r.Log.WithError(err).Warnf("Failed to complete remote OAuth2")
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
