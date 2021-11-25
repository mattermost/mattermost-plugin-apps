package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) initAdmin(rh *httpin.Handler) {
	rh.HandleFunc(path.UpdateAppListing,
		a.UpdateAppListing, httpin.RequireSysadminOrPlugin).Methods(http.MethodPost)
	rh.HandleFunc(path.InstallApp,
		a.InstallApp, httpin.RequireSysadminOrPlugin).Methods(http.MethodPost)
	rh.HandleFunc(path.EnableApp,
		a.EnableApp, httpin.RequireSysadminOrPlugin).Methods(http.MethodPost)
	rh.HandleFunc(path.DisableApp,
		a.DisableApp, httpin.RequireSysadminOrPlugin).Methods(http.MethodPost)
	rh.HandleFunc(path.UninstallApp,
		a.UninstallApp, httpin.RequireSysadminOrPlugin).Methods(http.MethodPost)
}

// UpdateAppListing adds (or updates) the specified Manifest to the local
// manifest store, making the App installable. The resulting listed manifest
// will combine the deployment information from the prior listing, and the new
// manifests as follows:
//   1. The "core" manifest (except Deploy) is updated to the new values.
//   2. Deploy types from the previously listed manifest are updated from the new manifest, or preserved.
//   3. Deploy types specified in "add_deploys" are copied from the new manifest.
//   4. "remove"
//   Path: /api/v1/add-listed-app
//   Method: POST
//   Input: JSON{
//      Manifest...
//      "add_deploys": []string e.g. ["aws_lambda","http"]
//      "remove_deploys": []string e.g. ["aws_lambda","http"]
//   Output: The updated listing manifest
func (a *restapi) UpdateAppListing(req *incoming.Request, w http.ResponseWriter, r *http.Request) {
	listReq := appclient.UpdateAppListingRequest{}
	err := json.NewDecoder(r.Body).Decode(&listReq)
	if err != nil {
		httputils.WriteError(w, utils.NewInvalidError(err, "failed to unmarshal input"))
		return
	}
	m, err := a.proxy.UpdateAppListing(req, listReq)
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
func (a *restapi) InstallApp(req *incoming.Request, w http.ResponseWriter, r *http.Request) {
	var input apps.App
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		httputils.WriteError(w, errors.Wrap(err, "failed to unmarshal input"))
		return
	}

	req.SetAppID(input.AppID)

	_, _, err = a.proxy.InstallApp(req, apps.Context{}, input.AppID, input.DeployType, false, "")
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
func (a *restapi) EnableApp(req *incoming.Request, w http.ResponseWriter, r *http.Request) {
	var input apps.App
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		httputils.WriteError(w, errors.Wrap(err, "failed to unmarshal input"))
		return
	}

	req.SetAppID(input.AppID)

	_, err = a.proxy.EnableApp(req, apps.Context{}, input.AppID)
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
func (a *restapi) DisableApp(req *incoming.Request, w http.ResponseWriter, r *http.Request) {
	var input apps.App
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		httputils.WriteError(w, errors.Wrap(err, "failed to unmarshal input"))
		return
	}

	req.SetAppID(input.AppID)

	_, err = a.proxy.DisableApp(req, apps.Context{}, input.AppID)
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
func (a *restapi) UninstallApp(req *incoming.Request, w http.ResponseWriter, r *http.Request) {
	var input apps.App
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		httputils.WriteError(w, errors.Wrap(err, "failed to unmarshal input"))
		return
	}

	req.SetAppID(input.AppID)

	_, err = a.proxy.UninstallApp(req, apps.Context{}, input.AppID)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}

func (a *restapi) initGetApp(rh *httpin.Handler) {
	rh = rh.PathPrefix(path.Apps)
	rh = rh.PathPrefix(`/{appid:[A-Za-z0-9-_.]+}`)
	rh.HandleFunc("",
		a.GetApp).Methods(http.MethodGet)
}

// GetApp returns the App's record.
//   Path: /apps/{AppID}
//   Method: GET
//   Input: none
//   Output: App
func (a *restapi) GetApp(req *incoming.Request, w http.ResponseWriter, r *http.Request) {
	appID := appIDVar(r)
	if appID == "" {
		httputils.WriteError(w, utils.NewInvalidError("app is required"))
		return
	}
	req.SetAppID(appID)

	app, err := a.proxy.GetInstalledApp(req, appID)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
	_ = httputils.WriteJSON(w, app)
}
