package request

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type check func(*Context, http.ResponseWriter, *http.Request) bool // check return true if the it was successful

type contextHandlerFunc func(c *Context, w http.ResponseWriter, r *http.Request)

type ContextHandler struct {
	handler contextHandlerFunc
	context *Context
	checks  []check
}

func AddContext(handler contextHandlerFunc, c *Context) *ContextHandler {
	return &ContextHandler{
		handler: handler,
		context: c,
	}
}

func (h *ContextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	context := h.context.Clone()
	context.RequestID = model.NewId()
	context.Log = context.Log.With(
		"path", r.URL.Path,
		"request", context.RequestID,
	)

	for _, check := range h.checks {
		succeeded := check(context, w, r)
		if !succeeded {
			return
		}
	}

	h.handler(context, w, r)
}

func getUserID(r *http.Request) string {
	return r.Header.Get(config.MattermostUserIDHeader)
}

func (h *ContextHandler) RequireUser() *ContextHandler {
	h.checks = append(h.checks, checkUser)

	return h
}

func checkUser(c *Context, w http.ResponseWriter, r *http.Request) bool {
	actingUserID := getUserID(r)
	if actingUserID == "" {
		httputils.WriteError(w, errors.Wrap(utils.ErrUnauthorized, "user ID is required"))
		return false
	}

	c.SetActingUserID(actingUserID)

	return true
}

func (h *ContextHandler) RequireSysadmin() *ContextHandler {
	h.checks = append(h.checks, checkSysadmin)

	return h
}

func checkSysadmin(c *Context, w http.ResponseWriter, r *http.Request) bool {
	if successful := checkUser(c, w, r); !successful {
		return successful
	}

	err := utils.EnsureSysAdmin(c.mm, c.ActingUserID())
	if err != nil {
		httputils.WriteError(w, errors.Wrap(utils.ErrUnauthorized, "user is not a system admin"))
		return false
	}

	c.sysAdminChecked = true

	return true
}

func (h *ContextHandler) RequireSysadminOrPlugin() *ContextHandler {
	check := func(c *Context, w http.ResponseWriter, r *http.Request) bool {
		pluginID := r.Header.Get(config.MattermostPluginIDHeader)
		if pluginID != "" {
			return true
		}

		return checkSysadmin(c, w, r)
	}

	h.checks = append(h.checks, check)

	return h
}
