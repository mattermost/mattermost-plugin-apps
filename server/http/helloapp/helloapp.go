package helloapp

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/appmodel"
	"github.com/mattermost/mattermost-plugin-apps/server/client"
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"

	"github.com/gorilla/mux"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
)

type helloapp struct {
	mm           *pluginapi.Client
	configurator configurator.Service
}

func Init(router *mux.Router, apps *apps.Service) {
	a := helloapp{
		mm:           apps.Mattermost,
		configurator: apps.Config,
	}

	subrouter := router.PathPrefix(constants.HelloAppPath).Subrouter()
	subrouter.HandleFunc("/mattermost-app.json", a.handleManifest).Methods("GET")
	subrouter.HandleFunc(constants.HelloInstallCompletePath, a.handleInstallcomplete).Methods("POST")
}

func (h *helloapp) handleManifest(w http.ResponseWriter, req *http.Request) {
	conf := h.configurator.GetConfig()
	rootURL := conf.PluginURL + constants.HelloAppPath
	httputils.WriteJSON(w,
		appmodel.Manifest{
			AppID:       "hello",
			DisplayName: "Hallo სამყარო",
			Description: "Hallo სამყარო test app",
			RootURL:     rootURL,
			RequestedPermissions: []appmodel.PermissionType{
				appmodel.PermissionUserJoinedChannelNotification,
				appmodel.PermissionActAsUser,
				appmodel.PermissionActAsBot,
			},
			CallbackURL:        rootURL + "/oauth",
			Homepage:           rootURL,
			InstallCompleteURL: rootURL + constants.HelloInstallCompletePath,
		})
}

func (h *helloapp) handleInstallcomplete(w http.ResponseWriter, req *http.Request) {
	var body client.InstallCompleteBody
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		h.mm.Log.Error("error decoding body", "err", err.Error())
		return
	}

	h.mm.Log.Debug("Received values", "BotID", body.BotID, "BotAccessToken", body.BotAccessToken, "OAuthID", body.OAuthAppID, "OAuthSecret", body.OAuthAppSecret)
}
