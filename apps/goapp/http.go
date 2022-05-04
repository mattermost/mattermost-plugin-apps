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
	Init(app *App)
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
		copy := *app
		creq.App = &copy
		creq.App.Log = app.Log.With("path", creq.Path)

		cresp := h(creq)
		if cresp.Type == apps.CallResponseTypeError {
			creq.App.Log.WithError(cresp).Debugw("Call failed.")
		}
		_ = httputils.WriteJSON(w, cresp)

		app.Log.Debugw("Call", "request", creq, "response", cresp)
	})
}

func FormHandler(h func(CallRequest) (apps.Form, error)) HandlerFunc {
	return func(creq CallRequest) apps.CallResponse {
		f, err := h(creq)
		if err != nil {
			creq.App.Log.WithError(err).Infow("failed to respond with form")
			return apps.NewErrorResponse(err)
		}
		return apps.NewFormResponse(f)
	}
}

func LookupHandler(h func(CallRequest) []apps.SelectOption) HandlerFunc {
	return func(creq CallRequest) apps.CallResponse {
		opts := h(creq)
		return apps.NewLookupResponse(opts)
	}
}

func CallHandler(h func(CallRequest) (string, error)) HandlerFunc {
	return func(creq CallRequest) apps.CallResponse {
		text, err := h(creq)
		if err != nil {
			creq.App.Log.WithError(err).Infow("failed to process call")
			return apps.NewErrorResponse(err)
		}
		return apps.NewTextResponse(text)
	}
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
					". Please use `/apps connect` to connect your ServiceNow account."))
		}
		return h(creq)
	}
}

func (creq CallRequest) AppProxyURL(paths ...string) string {
	p := path.Join(append([]string{creq.Context.AppPath}, paths...)...)
	return creq.Context.MattermostSiteURL + p
}
