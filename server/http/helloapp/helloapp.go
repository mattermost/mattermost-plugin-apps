package helloapp

import (
	"net/http"

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
}

func (h *helloapp) handleManifest(w http.ResponseWriter, req *http.Request) {
	conf := h.configurator.GetConfig()

	httputils.WriteJSON(w,
		apps.Manifest{
			AppID:       "hello",
			DisplayName: "Hallo სამყარო",
			Description: "Hallo სამყარო test app",
			RootURL:     conf.PluginURL + constants.HelloAppPath,
			RequestedPermissions: []apps.PermissionType{
				apps.PermissionUserJoinedChannelNotification,
				apps.PermissionActAsUser,
				apps.PermissionActAsBot,
			},
		})
}
