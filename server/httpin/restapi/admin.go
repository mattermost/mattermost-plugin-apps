package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) initAdmin(api *mux.Router, mm *pluginapi.Client) {
	api.HandleFunc(path.StoreListedApp,
		proxy.RequireSysadminOrPlugin(mm, a.StoreListedApp)).Methods("POST")
	api.HandleFunc(path.InstallApp,
		proxy.RequireSysadminOrPlugin(mm, a.InstallApp)).Methods("POST")
	api.HandleFunc(path.EnableApp,
		proxy.RequireSysadminOrPlugin(mm, a.EnableApp)).Methods("POST")
	api.HandleFunc(path.DisableApp,
		proxy.RequireSysadminOrPlugin(mm, a.DisableApp)).Methods("POST")
	api.HandleFunc(path.UninstallApp,
		proxy.RequireSysadminOrPlugin(mm, a.UninstallApp)).Methods("POST")
}

// StoreListedApp adds (or updates) the specified Manifest to the local manifest
// store, making the App installable.
//   Path: /api/v1/add-listed-app
//   Method: POST
//   Input: Manifest
//   Output: None
func (a *restapi) StoreListedApp(w http.ResponseWriter, r *http.Request, in proxy.Incoming) {
	var m apps.Manifest
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		httputils.WriteError(w, errors.Wrap(err, "failed to unmarshal input"))
		return
	}
	_, err = a.proxy.StoreLocalManifest(m)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

// InstallApp installs an App that is already deployed, either locally or in the
// Marketplace (if applicable).
//   Path: /api/v1/install-app
//   Method: POST
//   Input: JSON {app_id, deploy_type}
//   Output: None
func (a *restapi) InstallApp(w http.ResponseWriter, r *http.Request, in proxy.Incoming) {
	var input apps.App
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		httputils.WriteError(w, errors.Wrap(err, "failed to unmarshal input"))
		return
	}

	_, _, err = a.proxy.InstallApp(in, apps.Context{}, input.AppID, input.DeployType, false, "")
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

// EnableApp enables an App .
//   Path: /api/v1/enable-app
//   Method: POST
//   Input: JSON {app_id}
//   Output: None
func (a *restapi) EnableApp(w http.ResponseWriter, r *http.Request, in proxy.Incoming) {
	var input apps.App
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		httputils.WriteError(w, errors.Wrap(err, "failed to unmarshal input"))
		return
	}
	_, err = a.proxy.EnableApp(in, apps.Context{}, input.AppID)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

// DisableApp disables an App .
//   Path: /api/v1/disable-app
//   Method: POST
//   Input: JSON {app_id}
//   Output: None
func (a *restapi) DisableApp(w http.ResponseWriter, r *http.Request, in proxy.Incoming) {
	var input apps.App
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		httputils.WriteError(w, errors.Wrap(err, "failed to unmarshal input"))
		return
	}
	_, err = a.proxy.DisableApp(in, apps.Context{}, input.AppID)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

// UninstallApp uninstalls an App .
//   Path: /api/v1/uninstall-app
//   Method: POST
//   Input: JSON {app_id}
//   Output: None
func (a *restapi) UninstallApp(w http.ResponseWriter, r *http.Request, in proxy.Incoming) {
	var input apps.App
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		httputils.WriteError(w, errors.Wrap(err, "failed to unmarshal input"))
		return
	}
	_, err = a.proxy.UninstallApp(in, apps.Context{}, input.AppID)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) initGetApp(main *mux.Router, mm *pluginapi.Client) {
	appsRouters := main.PathPrefix(path.Apps).Subrouter()
	appRouter := appsRouters.PathPrefix(`/{appid:[A-Za-z0-9-_.]+}`).Subrouter()
	appRouter.HandleFunc("",
		proxy.RequireSysadminOrPlugin(mm, a.GetApp)).Methods("GET")
}

// GetApp returns the App's record.
//   Path: /apps/{AppID}
//   Method: GET
//   Input: none
//   Output: App
func (a *restapi) GetApp(w http.ResponseWriter, r *http.Request, _ proxy.Incoming) {
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
