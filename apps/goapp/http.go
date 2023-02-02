package goapp

import (
	"encoding/json"
	"net/http"
	"path"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type HandlerFunc func(CallRequest) apps.CallResponse

type Requirer interface {
	RequireSystemAdmin() bool
	RequireConnectedUser() bool
}

type Initializer interface {
	Init(app *App) error
}

func (app *App) HandleCall(p string, h HandlerFunc) {
	app.Router.Path(p).HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		creq := CallRequest{
			GoContext: req.Context(),
		}
		err := json.NewDecoder(req.Body).Decode(&creq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		creq.App = app
		creq.Log = app.log.With(creq)

		cresp := h(creq)
		if cresp.Type == apps.CallResponseTypeError {
			creq.Log.WithError(cresp).Debugw("Call failed.")
		}
		_ = httputils.WriteJSON(w, cresp)

		creq.Log.With(cresp).Debugf("Call %s returned %s:", creq.Path, cresp.Type)
	})
}

func RequireAdmin(h HandlerFunc) HandlerFunc {
	return func(creq CallRequest) apps.CallResponse {
		if !creq.IsSystemAdmin() {
			return apps.NewErrorResponse(
				utils.NewUnauthorizedError("system administrator role is required to invoke " + creq.Path))
		}
		return h(creq)
	}
}

func RequireConnectedUser(h HandlerFunc) HandlerFunc {
	return func(creq CallRequest) apps.CallResponse {
		if !creq.IsConnectedUser() {
			return apps.NewErrorResponse(
				utils.NewUnauthorizedError("missing user record, required for " + creq.Path +
					". Please use `/apps connect` to connect your account."))
		}
		return h(creq)
	}
}

func (creq CallRequest) AppProxyURL(paths ...string) string {
	p := path.Join(append([]string{creq.Context.AppPath}, paths...)...)
	return creq.Context.MattermostSiteURL + p
}
