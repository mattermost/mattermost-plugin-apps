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

func (a *restapi) handleGetApp(w http.ResponseWriter, r *http.Request, _ proxy.Incoming) {
	appID := appIDVar(r)
	if appID == "" {
		httputils.WriteError(w, utils.NewInvalidError("app is required"))
		return
	}
	app, err := a.proxy.GetInstalledApp(appID)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	httputils.WriteJSON(w, app)
}

func (a *restapi) handleEnableApp(w http.ResponseWriter, r *http.Request, in proxy.Incoming) {
	appID := appIDVar(r)
	if appID == "" {
		httputils.WriteError(w, utils.NewInvalidError("app is required"))
		return
	}
	_, err := a.proxy.EnableApp(in, apps.Context{}, appID)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) handleDisableApp(w http.ResponseWriter, r *http.Request, in proxy.Incoming) {
	appID := appIDVar(r)
	if appID == "" {
		httputils.WriteError(w, utils.NewInvalidError("app is required"))
		return
	}
	_, err := a.proxy.DisableApp(in, apps.Context{}, appID)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) handleInstallApp(w http.ResponseWriter, r *http.Request, in proxy.Incoming) {
	var m apps.Manifest
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		httputils.WriteError(w, errors.Wrap(err, "failed to unmarshal manifest"))
		return
	}

	_, err = a.proxy.AddLocalManifest(m)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	_, _, err = a.proxy.InstallApp(in, apps.Context{}, m.AppID, false, "")
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) handleUninstallApp(w http.ResponseWriter, r *http.Request, in proxy.Incoming) {
	appID := appIDVar(r)
	if appID == "" {
		httputils.WriteError(w, utils.NewInvalidError("app is required"))
		return
	}
	_, err := a.proxy.UninstallApp(in, apps.Context{}, appID)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}
