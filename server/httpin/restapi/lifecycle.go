package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
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
	actingUserID := r.URL.Query().Get("acting_user_id")

	cc := &apps.Context{
		ActingUserID: actingUserID,
	}

	cc = a.conf.GetConfig().SetContextDefaults(cc)

	_, err = a.proxy.AddLocalManifest(actingUserID, &m)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	_, _, err = a.proxy.InstallApp(true, sessionID, actingUserID, cc, false, "")
	if err != nil {
		httputils.WriteError(w, err)
		return
	}
}
