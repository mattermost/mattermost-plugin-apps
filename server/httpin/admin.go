package httpin

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

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
func (s *Service) UpdateAppListing(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	var err error
	defer func() { httputils.WriteErrorIfNeeded(w, err) }()

	listReq := appclient.UpdateAppListingRequest{}
	if err = json.NewDecoder(req.Body).Decode(&listReq); err != nil {
		err = utils.NewInvalidError(err, "failed to unmarshal incoming request")
		return
	}
	m, err := s.Proxy.UpdateAppListing(r, listReq)
	if err != nil {
		return
	}

	_ = httputils.WriteJSON(w, m)
}

// InstallApp installs an App that is already deployed, either locally or in the
// Marketplace (if applicable).
//   Path: /api/v1/install-app
//   Method: POST
//   Input: JSON {app_id, deploy_type}
//   Output: JSON, unsanitized App record
func (s *Service) InstallApp(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	var err error
	defer func() { httputils.WriteErrorIfNeeded(w, err) }()

	var input apps.App
	if err = json.NewDecoder(req.Body).Decode(&input); err != nil {
		err = utils.NewInvalidError(err, "failed to unmarshal incoming request")
		return
	}
	app, _, err := s.Proxy.InstallApp(r, apps.Context{}, input.AppID, input.DeployType, false, "")
	if err != nil {
		return
	}
	_ = httputils.WriteJSON(w, app)
}

// EnableApp enables an App .
//   Path: /api/v1/enable-app
//   Method: POST
//   Input: JSON {app_id}
//   Output: none
func (s *Service) EnableApp(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	var err error
	defer func() { httputils.WriteErrorIfNeeded(w, err) }()

	var input apps.App
	if err = json.NewDecoder(req.Body).Decode(&input); err != nil {
		err = utils.NewInvalidError(err, "failed to unmarshal incoming request")
		return
	}
	text, err := s.Proxy.EnableApp(r, apps.Context{}, input.AppID)
	if err != nil {
		return
	}
	_, _ = w.Write([]byte(text))
}

// DisableApp disables an App .
//   Path: /api/v1/disable-app
//   Method: POST
//   Input: JSON {app_id}
//   Output: JSON, unsanitized App record
func (s *Service) DisableApp(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	var err error
	defer func() { httputils.WriteErrorIfNeeded(w, err) }()

	var input apps.App
	if err = json.NewDecoder(req.Body).Decode(&input); err != nil {
		err = utils.NewInvalidError(err, "failed to unmarshal incoming request")
		return
	}
	text, err := s.Proxy.DisableApp(r, apps.Context{}, input.AppID)
	if err != nil {
		return
	}
	_, _ = w.Write([]byte(text))
}

// UninstallApp uninstalls an App .
//   Path: /api/v1/uninstall-app
//   Method: POST
//   Input: JSON {app_id}
//   Output: None
func (s *Service) UninstallApp(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	var err error
	defer func() { httputils.WriteErrorIfNeeded(w, err) }()

	var input apps.App
	if err = json.NewDecoder(req.Body).Decode(&input); err != nil {
		err = utils.NewInvalidError(err, "failed to unmarshal incoming request")
		return
	}
	if _, err = s.Proxy.UninstallApp(r, apps.Context{}, input.AppID, false); err != nil {
		return
	}
}

// GetApp returns the App's record. If requestor is a system administrator, the
// raw record with secrets is returned, otherwise the output is sanitized.
//   Path: /apps/{AppID}
//   Method: GET
//   Input: none
//   Output: App
func (s *Service) GetApp(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	var err error
	defer func() { httputils.WriteErrorIfNeeded(w, err) }()

	app, err := s.Proxy.GetApp(r)
	if err != nil {
		return
	}
	_ = httputils.WriteJSON(w, app)
}

func (s *Service) GetMarketplace(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	filter := req.URL.Query().Get("filter")
	includePlugins := req.URL.Query().Get("include_plugins") != ""

	result := s.Proxy.GetListedApps(filter, includePlugins)
	_ = httputils.WriteJSON(w, result)
}
