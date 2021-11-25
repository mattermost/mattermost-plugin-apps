package incoming

import (
	"context"
	"net/http"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
	"github.com/mattermost/mattermost-plugin-apps/utils/sessionutils"
)

type check func(*Request, http.ResponseWriter, *http.Request) bool // check return true if the it was successful

type contextHandlerFunc func(c *Request, w http.ResponseWriter, r *http.Request)

type RequestHandler struct {
	handler contextHandlerFunc
	context *Request
	checks  []check
}

func AddContext(handler contextHandlerFunc, c *Request) *RequestHandler {
	return &RequestHandler{
		handler: handler,
		context: c,
	}
}

func (h *RequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := h.context.Clone()
	c.requestID = model.NewId()
	c.Log = c.Log.With(
		"path", r.URL.Path,
		"request", c.requestID,
	)
	ctx, cancel := context.WithTimeout(r.Context(), config.RequestTimeout)
	defer cancel()
	c.ctx = ctx

	for _, check := range h.checks {
		succeeded := check(c, w, r)
		if !succeeded {
			return
		}
	}

	h.handler(c, w, r)
}

func getUserID(r *http.Request) string {
	return r.Header.Get(config.MattermostUserIDHeader)
}

func (h *RequestHandler) RequireUser() *RequestHandler {
	h.checks = append(h.checks, checkUser)

	return h
}

func checkUser(c *Request, w http.ResponseWriter, r *http.Request) bool {
	actingUserID := getUserID(r)
	if actingUserID == "" {
		httputils.WriteError(w, utils.NewUnauthorizedError("user ID is required"))
		return false
	}

	c.SetActingUserID(actingUserID)

	return true
}

func (h *RequestHandler) RequireSysadmin() *RequestHandler {
	h.checks = append(h.checks, checkSysadmin)

	return h
}

func checkSysadmin(c *Request, w http.ResponseWriter, r *http.Request) bool {
	if c.sysAdminChecked {
		return true
	}

	if successful := checkUser(c, w, r); !successful {
		return false
	}

	if !c.mm.User.HasPermissionTo(c.ActingUserID(), model.PermissionManageSystem) {
		httputils.WriteError(w, utils.NewUnauthorizedError("user is not a system admin"))
		return false
	}

	c.sysAdminChecked = true

	return true
}

func checkPlugin(c *Request, w http.ResponseWriter, r *http.Request) bool {
	pluginID := r.Header.Get(config.MattermostPluginIDHeader)
	return pluginID != ""
}

func (h *RequestHandler) RequireSysadminOrPlugin() *RequestHandler {
	check := func(c *Request, w http.ResponseWriter, r *http.Request) bool {
		if checkPlugin(c, w, r) {
			return true
		}

		return checkSysadmin(c, w, r)
	}

	h.checks = append(h.checks, check)

	return h
}

func checkApp(c *Request, w http.ResponseWriter, r *http.Request) bool {
	sessionID := r.Header.Get(config.MattermostSessionIDHeader)
	if sessionID == "" {
		httputils.WriteError(w, utils.NewUnauthorizedError("a session is required"))
		return false
	}

	s, err := c.mm.Session.Get(sessionID)
	if err != nil {
		httputils.WriteError(w, errors.New("session check failed"))
		return false
	}

	appID := sessionutils.GetAppID(s)
	if appID == "" {
		httputils.WriteError(w, utils.NewUnauthorizedError("not an app session"))
		return false
	}

	c.SetAppID(appID)

	return true
}

func (h *RequestHandler) RequireApp() *RequestHandler {
	h.checks = append(h.checks, checkApp)

	return h
}
