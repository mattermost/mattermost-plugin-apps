package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) handleInstallApp(w http.ResponseWriter, r *http.Request) {
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

	_, _, err = a.proxy.InstallApp(sessionID, actingUserID, cc, false, "")
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}
