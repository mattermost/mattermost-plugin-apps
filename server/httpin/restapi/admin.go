package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) initAdmin(api *mux.Router, mm *pluginapi.Client) {
	api.HandleFunc(path.UpdateAppListing,
		proxy.RequireSysadminOrPlugin(mm, a.UpdateAppListing)).Methods("POST")
	api.HandleFunc(path.InstallApp,
		proxy.RequireSysadminOrPlugin(mm, a.InstallApp)).Methods("POST")
	api.HandleFunc(path.EnableApp,
		proxy.RequireSysadminOrPlugin(mm, a.EnableApp)).Methods("POST")
	api.HandleFunc(path.DisableApp,
		proxy.RequireSysadminOrPlugin(mm, a.DisableApp)).Methods("POST")
	api.HandleFunc(path.UninstallApp,
		proxy.RequireSysadminOrPlugin(mm, a.UninstallApp)).Methods("POST")
}

// UpdateAppListing adds (or updates) the specified Manifest to the local
// manifest store, making the App installable. The resulting listed manifest
// will combine the deployment information from the prior listing, and the new
// manifests as follows:
//   1. The "core" manifest (except Deploy) is updated to the new values.
//   2. Deployment types from the previously listed manifest are updated from the new manifest, or preserved.
//   3. Deployment types specified in "add_deployments" are copied from the new manifest.
//   4. "remove"
//   Path: /api/v1/add-listed-app
//   Method: POST
//   Input: JSON{
//      Manifest...
//      "add_deployments": []string e.g. ["aws_lambda","http"]
//      "remove_deployments": []string e.g. ["aws_lambda","http"]
//   Output: The updated listing manifest
func (a *restapi) UpdateAppListing(w http.ResponseWriter, r *http.Request, in proxy.Incoming) {
	req := appclient.UpdateAppListingRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httputils.WriteError(w, utils.NewInvalidError(err, "failed to unmarshal input"))
		return
	}
	m, err := a.proxy.UpdateAppListing(req)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	_ = httputils.WriteJSON(w, m)
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
	_ = httputils.WriteJSON(w, app)
}
