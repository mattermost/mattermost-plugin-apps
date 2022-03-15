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

type check func(req *incoming.Request, mm *pluginapi.Client, w http.ResponseWriter, r *http.Request) bool // check return true if the it was successful

type handlerFunc func(req *incoming.Request, w http.ResponseWriter, r *http.Request)

type Handler struct {
	mm             *pluginapi.Client
	config         config.Service
	log            utils.Logger
	sessionService incoming.SessionService
	router         *mux.Router

	handlerFunc handlerFunc
	checks      []check
}

func NewHandler(mm *pluginapi.Client, config config.Service, log utils.Logger, session incoming.SessionService, router *mux.Router) *Handler {
	rh := &Handler{
		mm:             mm,
		config:         config,
		log:            log,
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
		log:            h.log,
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

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), config.RequestTimeout)
	defer cancel()
	req := incoming.NewRequest(h.mm, h.config, h.log, h.sessionService, incoming.WithCtx(ctx))

	req.Log = req.Log.With(
		"path", r.URL.Path,
	)

	defer func() {
		if x := recover(); x != nil {
			stack := string(debug.Stack())

			req.Log.Errorw(
				"Recovered from a panic in an HTTP handler",
				"url", r.URL.String(),
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
		succeeded := check(req, h.mm, w, r)
		if !succeeded {
			return
		}
	}

	h.handlerFunc(req, w, r)
}

func getUserID(r *http.Request) string {
	return r.Header.Get(config.MattermostUserIDHeader)
}

func RequireUser(req *incoming.Request, mm *pluginapi.Client, w http.ResponseWriter, r *http.Request) bool {
	actingUserID := getUserID(r)
	if actingUserID == "" {
		httputils.WriteError(w, utils.NewUnauthorizedError("user ID is required"))
		return false
	}

	req.SetActingUserID(actingUserID)

	return true
}

func RequireSysadmin(req *incoming.Request, mm *pluginapi.Client, w http.ResponseWriter, r *http.Request) bool {
	if successful := RequireUser(req, mm, w, r); !successful {
		return false
	}

	if !mm.User.HasPermissionTo(req.ActingUserID(), model.PermissionManageSystem) {
		httputils.WriteError(w, utils.NewUnauthorizedError("user is not a system admin"))
		return false
	}

	return true
}

func RequireSysadminOrPlugin(req *incoming.Request, mm *pluginapi.Client, w http.ResponseWriter, r *http.Request) bool {
	pluginID := r.Header.Get(config.MattermostPluginIDHeader)
	if pluginID != "" {
		return true
	}

	return RequireSysadmin(req, mm, w, r)
}

func RequireApp(req *incoming.Request, mm *pluginapi.Client, w http.ResponseWriter, r *http.Request) bool {
	sessionID := r.Header.Get(config.MattermostSessionIDHeader)
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

	req.SetAppID(appID)

	return true
}
