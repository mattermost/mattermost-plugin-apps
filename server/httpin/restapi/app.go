package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) handleGetApp(w http.ResponseWriter, r *http.Request, pluginID, _, actingUserID string) {
	// Only check non-plugin requests
	if pluginID == "" {
		err := utils.EnsureSysAdmin(a.mm, actingUserID)
		if err != nil {
			httputils.WriteError(w, errors.Wrap(err, "only admins can get apps"))
			return
		}
	}

	appID := appIDVar(r)
	if appID == "" {
		httputils.WriteError(w, errors.Wrap(utils.ErrInvalid, "app is required"))
		return
	}

	app, err := a.proxy.GetInstalledApp(appID)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	httputils.WriteJSON(w, app)
}

func (a *restapi) handleEnableApp(w http.ResponseWriter, r *http.Request, pluginID, sessionID, actingUserID string) {
	// Only check non-plugin requests
	if pluginID == "" {
		err := utils.EnsureSysAdmin(a.mm, actingUserID)
		if err != nil {
			httputils.WriteError(w, errors.Wrap(err, "only admins can get apps"))
			return
		}
	}

	appID := appIDVar(r)
	if appID == "" {
		httputils.WriteError(w, errors.Wrap(utils.ErrInvalid, "app is required"))
		return
	}

	var client proxy.MMClient
	if pluginID != "" {
		client = a.proxy.GetMMRPCClient()
	} else {
		var err error
		client, err = a.proxy.GetMMHTTPClient(sessionID, actingUserID)
		if err != nil {
			httputils.WriteError(w, errors.Wrap(utils.ErrInvalid, "invalid session"))
			return
		}
	}

	cc := &apps.Context{
		ActingUserID: actingUserID,
		UserID:       actingUserID,
	}
	cc = a.conf.GetConfig().SetContextDefaultsForApp(appID, cc)

	_, err := a.proxy.EnableApp(client, sessionID, cc, appID)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) handleDisableApp(w http.ResponseWriter, r *http.Request, pluginID, sessionID, actingUserID string) {
	// Only check non-plugin requests
	if pluginID == "" {
		err := utils.EnsureSysAdmin(a.mm, actingUserID)
		if err != nil {
			httputils.WriteError(w, errors.Wrap(err, "only admins can get apps"))
			return
		}
	}

	appID := appIDVar(r)
	if appID == "" {
		httputils.WriteError(w, errors.Wrap(utils.ErrInvalid, "app is required"))
		return
	}

	var client proxy.MMClient
	if pluginID != "" {
		client = a.proxy.GetMMRPCClient()
	} else {
		var err error
		client, err = a.proxy.GetMMHTTPClient(sessionID, actingUserID)
		if err != nil {
			httputils.WriteError(w, errors.Wrap(utils.ErrInvalid, "invalid session"))
			return
		}
	}

	cc := &apps.Context{
		ActingUserID: actingUserID,
		UserID:       actingUserID,
	}
	cc = a.conf.GetConfig().SetContextDefaultsForApp(appID, cc)

	_, err := a.proxy.DisableApp(client, sessionID, cc, appID)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) handleInstallApp(w http.ResponseWriter, r *http.Request, pluginID, sessionID, actingUserID string) {
	// Only check non-plugin requests
	if pluginID == "" {
		err := utils.EnsureSysAdmin(a.mm, actingUserID)
		if err != nil {
			httputils.WriteError(w, errors.Wrap(err, "only admins can get apps"))
			return
		}
	}

	var m apps.Manifest
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		httputils.WriteError(w, errors.Wrap(err, "failed to unmarshal manifest"))
		return
	}

	var client proxy.MMClient
	if pluginID != "" {
		client = a.proxy.GetMMRPCClient()
	} else {
		client, err = a.proxy.GetMMHTTPClient(sessionID, actingUserID)
		if err != nil {
			httputils.WriteError(w, errors.Wrap(utils.ErrInvalid, "invalid session"))
			return
		}
	}

	cc := &apps.Context{
		ActingUserID: actingUserID,
		UserID:       actingUserID,
	}
	cc = a.conf.GetConfig().SetContextDefaultsForApp(m.AppID, cc)

	_, err = a.proxy.AddLocalManifest(actingUserID, &m)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	_, _, err = a.proxy.InstallApp(client, sessionID, cc, false, "", pluginID)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) handleUninstallApp(w http.ResponseWriter, r *http.Request, pluginID, sessionID, actingUserID string) {
	// Only check non-plugin requests
	if pluginID == "" {
		err := utils.EnsureSysAdmin(a.mm, actingUserID)
		if err != nil {
			httputils.WriteError(w, errors.Wrap(err, "only admins can get apps"))
			return
		}
	}

	appID := appIDVar(r)
	if appID == "" {
		httputils.WriteError(w, errors.Wrap(utils.ErrInvalid, "app is required"))
		return
	}

	var client proxy.MMClient
	if pluginID != "" {
		client = a.proxy.GetMMRPCClient()
	} else {
		var err error
		client, err = a.proxy.GetMMHTTPClient(sessionID, actingUserID)
		if err != nil {
			httputils.WriteError(w, errors.Wrap(utils.ErrInvalid, "invalid session"))
			return
		}
	}

	cc := &apps.Context{
		ActingUserID: actingUserID,
		UserID:       actingUserID,
	}
	cc = a.conf.GetConfig().SetContextDefaultsForApp(appID, cc)

	_, err := a.proxy.UninstallApp(client, sessionID, cc, appID)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}
