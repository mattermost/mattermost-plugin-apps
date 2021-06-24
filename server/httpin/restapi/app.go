package restapi

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
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

	cc := &apps.Context{
		ActingUserID: actingUserID,
	}
	cc = a.conf.GetConfig().SetContextDefaultsForApp(appID, cc)

	_, err := a.proxy.EnableApp(sessionID, actingUserID, cc, appID)
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
	log.Printf("disabling: %#+v\n", appID)

	cc := &apps.Context{
		ActingUserID: actingUserID,
	}
	cc = a.conf.GetConfig().SetContextDefaultsForApp(appID, cc)

	_, err := a.proxy.DisableApp(sessionID, actingUserID, cc, appID)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) handleInstallApp(w http.ResponseWriter, r *http.Request, pluginID string) {
	var m apps.Manifest
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		httputils.WriteError(w, errors.Wrap(err, "failed to unmarshal manifest"))
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		httputils.WriteError(w, utils.NewInvalidError("session_id secret was not provided"))
		return
	}

	actingUserID := r.URL.Query().Get("acting_user_id")
	if actingUserID == "" {
		httputils.WriteError(w, utils.NewInvalidError("acting_user_id secret was not provided"))
		return
	}

	cc := &apps.Context{
		ActingUserID: actingUserID,
	}
	cc = a.conf.GetConfig().SetContextDefaultsForApp(m.AppID, cc)

	_, err = a.proxy.AddLocalManifest(actingUserID, &m)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	_, _, err = a.proxy.InstallApp(sessionID, actingUserID, cc, false, "", pluginID)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}
