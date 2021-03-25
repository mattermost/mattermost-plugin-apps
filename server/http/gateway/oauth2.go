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

	creq := apps.CallRequest{
		Call: apps.Call{
			Path: "/oauth2/redirect", // <>/<>
		},
		Type: apps.CallTypeSubmit,
		Context: g.conf.GetConfig().SetContextDefaultsForApp(
			&apps.Context{
				ActingUserID: actingUserID,
			},
			apps.AppID(appID),
		),
	}
	cresp := g.proxy.Call(apps.SessionToken(token), &creq)
	if cresp.Type == apps.CallResponseTypeError {
		httputils.WriteInternalServerError(w, cresp)
		return
	}
	if cresp.Type != "" && cresp.Type != apps.CallResponseTypeOK {
		httputils.WriteInternalServerError(w, errors.Errorf("oauth2: unexpected response type from the app: %q", cresp.Type))
		return
	}
	redirectURL, ok := cresp.Data.(string)
	if !ok {
		httputils.WriteInternalServerError(w, errors.Errorf("oauth2: unexpected data type from the app: %T, expected string (redirect URL)", cresp.Data))
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

	creq := apps.CallRequest{
		Call: apps.Call{
			Path: "/oauth2/complete", // <>/<>
		},
		Type: apps.CallTypeSubmit,
		Context: g.conf.GetConfig().SetContextDefaultsForApp(
			&apps.Context{
				ActingUserID: actingUserID,
			},
			apps.AppID(appID),
		),
		Values: map[string]interface{}{
			"q": req.URL.RawQuery,
		},
	}
	cresp := g.proxy.Call(apps.SessionToken(token), &creq)
	if cresp.Type == apps.CallResponseTypeError {
		httputils.WriteInternalServerError(w, cresp)
		return
	}
	if cresp.Type != "" && cresp.Type != apps.CallResponseTypeOK {
		httputils.WriteInternalServerError(w, errors.Errorf("oauth2: unexpected response type from the app: %q", cresp.Type))
		return
	}

	w.Header().Set("Content-Type", "text/html")
	html := `
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
	`
	_, _ = w.Write([]byte(html))
}
