package httpin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gorilla/mux"
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
	"github.com/mattermost/mattermost-plugin-apps/utils/sessionutils"
)

type check func(*incoming.Request, *pluginapi.Client, http.ResponseWriter, *http.Request) bool // check return true if the it was successful

type handlerFunc func(*incoming.Request, http.ResponseWriter, *http.Request)

type Handler struct {
	mm             *pluginapi.Client
	config         config.Service
	baseLog        utils.Logger
	sessionService incoming.SessionService
	router         *mux.Router

	handlerFunc handlerFunc
	checks      []check
}

func NewHandler(mm *pluginapi.Client, config config.Service, baseLog utils.Logger, session incoming.SessionService, router *mux.Router) *Handler {
	rh := &Handler{
		mm:             mm,
		config:         config,
		baseLog:        baseLog,
		sessionService: session,
		router:         router,
	}

	return rh
}

// clone creates a shallow copy of Handler, allowing clones to apply changes per handler func.
func (h *Handler) clone() *Handler {
	return &Handler{
		mm:             h.mm,
		config:         h.config,
		baseLog:        h.baseLog,
		sessionService: h.sessionService,
		router:         h.router,

		// Don't copy the following fields as they are specific to the handler func
		// - handler
		// - checks
	}
}

func (h *Handler) PathPrefix(tpl string) *Handler {
	clone := h.clone()

	clone.router = clone.router.PathPrefix(tpl).Subrouter()

	return clone
}

func (h *Handler) HandleFunc(path string, handlerFunc handlerFunc, checks ...check) *mux.Route {
	clone := h.clone()

	clone.checks = checks
	clone.handlerFunc = handlerFunc

	return clone.router.Handle(path, clone)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), config.RequestTimeout)
	defer cancel()
	r := incoming.NewRequest(h.mm, h.config, h.baseLog, h.sessionService, incoming.WithCtx(ctx))

	// TODO: what else to add? 1/5 add clones of CallResponse to Request and log it.
	r.Log = r.Log.With(
		"path", req.URL.Path,
	)

	defer func() {
		if x := recover(); x != nil {
			stack := string(debug.Stack())

			r.Log.Errorw(
				"Recovered from a panic in an HTTP handler",
				"url", req.URL.String(),
				"error", x,
				"stack", string(debug.Stack()),
			)

			txt := "Paniced while handling the request. "

			if h.config.Get().DeveloperMode {
				txt += fmt.Sprintf("Error: %v. Stack: %v", x, stack)
			} else {
				txt += "Please check the server logs for more details."
			}

			http.Error(w, txt, http.StatusInternalServerError)
		}
	}()

	for _, check := range h.checks {
		succeeded := check(r, h.mm, w, req)
		if !succeeded {
			return
		}
	}

	h.handlerFunc(r, w, req)

	if h.config.Get().DeveloperMode {
		r.Log.Debugf("HTTP")
	}
}

func getUserID(req *http.Request) string {
	return req.Header.Get(config.MattermostUserIDHeader)
}

func RequireUser(r *incoming.Request, mm *pluginapi.Client, w http.ResponseWriter, req *http.Request) bool {
	actingUserID := getUserID(req)
	if actingUserID == "" {
		httputils.WriteError(w, utils.NewUnauthorizedError("user ID is required"))
		return false
	}

	r.SetActingUserID(actingUserID)

	return true
}

func RequireSysadmin(r *incoming.Request, mm *pluginapi.Client, w http.ResponseWriter, req *http.Request) bool {
	if successful := RequireUser(r, mm, w, req); !successful {
		return false
	}

	if !mm.User.HasPermissionTo(r.ActingUserID(), model.PermissionManageSystem) {
		httputils.WriteError(w, utils.NewUnauthorizedError("user is not a system admin"))
		return false
	}

	return true
}

func RequireSysadminOrPlugin(r *incoming.Request, mm *pluginapi.Client, w http.ResponseWriter, req *http.Request) bool {
	pluginID := req.Header.Get(config.MattermostPluginIDHeader)
	if pluginID != "" {
		return true
	}

	return RequireSysadmin(r, mm, w, req)
}

func RequireApp(r *incoming.Request, mm *pluginapi.Client, w http.ResponseWriter, req *http.Request) bool {
	sessionID := req.Header.Get(config.MattermostSessionIDHeader)
	if sessionID == "" {
		httputils.WriteError(w, utils.NewUnauthorizedError("a session is required"))
		return false
	}

	s, err := mm.Session.Get(sessionID)
	if err != nil {
		httputils.WriteError(w, errors.New("session check failed"))
		return false
	}

	appID := sessionutils.GetAppID(s)
	if appID == "" {
		httputils.WriteError(w, utils.NewUnauthorizedError("not an app session"))
		return false
	}

	r.SetAppID(appID)

	return true
}
