package request

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type check func(w http.ResponseWriter, r *http.Request) bool // check return true if the it was successful

type contextHandlerFunc func(c *Context, w http.ResponseWriter, r *http.Request)

type contextHandler struct {
	handler contextHandlerFunc
	context *Context
	checks  []check
}

func AddContext(handler contextHandlerFunc, c *Context) *contextHandler {
	return &contextHandler{
		handler: handler,
		context: c,
	}
}

func (h *contextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	context := h.context.Clone()
	context.RequestID = model.NewId()
	context.Log = context.Log.With(
		"path", r.URL.Path,
		"request", context.RequestID,
	)

	for _, check := range h.checks {
		succeeded := check(w, r)
		if !succeeded {
			return
		}
	}

	h.handler(h.context, w, r)
}

func getUserID(r *http.Request) string {
	return r.Header.Get(config.MattermostUserIDHeader)
}

func (h *contextHandler) RequireUser() *contextHandler {
	h.checks = append(h.checks, h.checkUser)

	return h
}

func (h *contextHandler) checkUser(w http.ResponseWriter, r *http.Request) bool {
	actingUserID := getUserID(r)
	if actingUserID == "" {
		httputils.WriteError(w, errors.Wrap(utils.ErrUnauthorized, "user ID is required"))
		return false
	}

	h.context.SetActingUserID(actingUserID)

	return true
}

func (h *contextHandler) RequireSysadmin() *contextHandler {
	h.checks = append(h.checks, h.checkSysadmin)

	return h
}

func (h *contextHandler) checkSysadmin(w http.ResponseWriter, r *http.Request) bool {
	if successful := h.checkUser(w, r); !successful {
		return successful
	}

	err := utils.EnsureSysAdmin(h.context.mm, h.context.ActingUserID())
	if err != nil {
		httputils.WriteError(w, errors.Wrap(utils.ErrUnauthorized, "user is not a system admin"))
		return false
	}

	h.context.sysAdminChecked = true

	return true
}

func (h *contextHandler) RequireSysadminOrPlugin() *contextHandler {
	check := func(w http.ResponseWriter, r *http.Request) bool {
		pluginID := r.Header.Get(config.MattermostPluginIDHeader)
		if pluginID != "" {
			return true
		}

		return h.checkSysadmin(w, r)
	}

	h.checks = append(h.checks, check)

	return h
}
