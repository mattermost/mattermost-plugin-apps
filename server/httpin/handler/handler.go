package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
	"github.com/mattermost/mattermost-plugin-apps/utils/sessionutils"
)

const AppIDVar = "appid"
const AppIDPath = "/{appid}"

// check returns a potentially modified incoming.Request upon success. Upon
// failure it returns nil and writes the HTTP error.
type check func(*incoming.Request, http.ResponseWriter, *http.Request) *incoming.Request

type handlerFunc func(*incoming.Request, http.ResponseWriter, *http.Request)

type Handler struct {
	Config config.Service
	Proxy  proxy.Service

	baseLog     utils.Logger
	router      *mux.Router
	handlerFunc handlerFunc
	checks      []check
}

func NewHandler(proxy proxy.Service, config config.Service, baseLog utils.Logger) *Handler {
	h := &Handler{
		baseLog: baseLog,
		Config:  config,
		Proxy:   proxy,
		router:  mux.NewRouter(),
	}
	h.router.Handle("{anything:.*}", http.NotFoundHandler())
	return h
}

// clone creates a shallow copy of Handler, allowing clones to apply changes per
// handler func. Don't copy the following fields as they are specific to the
// handler func: handlerFunc, checks.
func (h *Handler) clone() *Handler {
	clone := *h
	clone.handlerFunc = nil
	clone.checks = nil
	return &clone
}

func (h *Handler) PathPrefix(prefix string) *Handler {
	clone := h.clone()
	clone.router = clone.router.PathPrefix(prefix).Subrouter()
	return clone
}

func (h *Handler) HandleFunc(path string, handlerFunc handlerFunc, checks ...check) *mux.Route {
	clone := h.clone()
	clone.checks = checks
	clone.handlerFunc = handlerFunc
	return clone.router.Handle(path, clone)
}

// ServePluginHTTP is the interface invoked from the plugin's ServeHTTP
func (h *Handler) ServePluginHTTP(c *plugin.Context, w http.ResponseWriter, req *http.Request) {
	req.Header.Set(config.MattermostSessionIDHeader, c.SessionId)
	h.router.ServeHTTP(w, req)
}

// ServeHTTP is the go http.Handler (mux compliant).
func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var cancel context.CancelFunc
	r := h.Proxy.NewIncomingRequest(h.baseLog).WithTimeout(config.RequestTimeout, &cancel)
	defer cancel()

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

			if h.Config.Get().DeveloperMode {
				txt += fmt.Sprintf("Error: %v. Stack: %v", x, stack)
			} else {
				txt += "Please check the server logs for more details."
			}

			http.Error(w, txt, http.StatusInternalServerError)
		}
	}()

	actingUserID := req.Header.Get(config.MattermostUserIDHeader)
	if actingUserID != "" {
		r = r.WithActingUserID(actingUserID)
	}
	pluginID := req.Header.Get(config.MattermostPluginIDHeader)
	if pluginID != "" {
		r = r.WithSourcePluginID(pluginID)
	}
	// SourceAppID is not set here, only in Require because it is more expensive.

	for _, check := range h.checks {
		r = check(r, w, req)
		if r == nil {
			return
		}
	}

	h.handlerFunc(r, w, req)

	if h.Config.Get().DeveloperMode {
		r.Log.Debugf("HTTP")
	}
}

func (h *Handler) error(txt string, r *incoming.Request, w http.ResponseWriter, req *http.Request) *incoming.Request {
	err := utils.NewUnauthorizedError(txt)
	r.Log = r.Log.WithError(err)
	httputils.WriteError(w, err)
	return nil
}

func (h *Handler) RequireActingUser(r *incoming.Request, w http.ResponseWriter, req *http.Request) *incoming.Request {
	if r.ActingUserID() == "" {
		return h.error("user ID is required", r, w, req)
	}
	return r
}

func (h *Handler) RequireSysadmin(r *incoming.Request, w http.ResponseWriter, req *http.Request) *incoming.Request {
	r = h.RequireActingUser(r, w, req)
	if r == nil {
		return nil
	}
	mm := r.Config().MattermostAPI()
	if !mm.User.HasPermissionTo(r.ActingUserID(), model.PermissionManageSystem) {
		return h.error("access to this operation is limited to system administrators", r, w, req)
	}
	return r
}

func (h *Handler) RequireSysadminOrPlugin(r *incoming.Request, w http.ResponseWriter, req *http.Request) *incoming.Request {
	radmin := h.RequireSysadmin(r, w, req)
	if radmin != nil {
		return radmin
	}
	if r.SourcePluginID() == "" {
		return h.error("access to this operation is limited to system administrators, or plugins", r, w, req)
	}
	return r
}

func (h *Handler) RequireFromApp(r *incoming.Request, w http.ResponseWriter, req *http.Request) *incoming.Request {
	sessionID := req.Header.Get(config.MattermostSessionIDHeader)
	if sessionID == "" {
		httputils.WriteError(w, utils.NewUnauthorizedError("a session is required"))
		return nil
	}
	mm := r.Config().MattermostAPI()
	s, err := mm.Session.Get(sessionID)
	if err != nil {
		httputils.WriteError(w, errors.New("session check failed"))
		return nil
	}
	appID := sessionutils.GetAppID(s)
	if appID == "" {
		httputils.WriteError(w, utils.NewUnauthorizedError("not an app session"))
		return nil
	}
	return r.WithSourceAppID(appID)
}

func (h *Handler) RequireToAppFromPath(r *incoming.Request, w http.ResponseWriter, req *http.Request) *incoming.Request {
	s, ok := mux.Vars(req)[AppIDVar]
	if !ok {
		httputils.WriteError(w, utils.NewUnauthorizedError("app ID is required in path"))
		return nil
	}
	return r.WithDestination(apps.AppID(s))
}
