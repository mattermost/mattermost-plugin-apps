package gateway

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (g *gateway) remoteOAuth2Redirect(w http.ResponseWriter, req *http.Request, actingUserID, token string) {
	vars := mux.Vars(req)

	appID := vars["app_id"]
	if appID == "" {
		httputils.WriteBadRequestError(w, errors.New("app_id not specified"))
		return
	}

	redirectURL, err := g.proxy.GetRemoteOAuth2RedirectURL(apps.AppID(appID), actingUserID, token)
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}

	http.Redirect(w, req, redirectURL, http.StatusTemporaryRedirect)
}

func (g *gateway) remoteOAuth2Complete(w http.ResponseWriter, req *http.Request, actingUserID, token string) {
	vars := mux.Vars(req)

	appID := vars["app_id"]
	if appID == "" {
		httputils.WriteBadRequestError(w, errors.New("app_id not specified"))
		return
	}

	q := req.URL.Query()
	urlValues := map[string]interface{}{}
	for key := range q {
		urlValues[key] = q.Get(key)
	}

	err := g.proxy.CompleteRemoteOAuth2(apps.AppID(appID), actingUserID, token, urlValues)
	if err != nil {
		httputils.WriteInternalServerError(w, err)
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
