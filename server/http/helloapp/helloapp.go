package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-api/experimental/oauther"
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

	OAuther           oauther.OAuther
	OAuthClientID     string
	OAuthClientSecret string
}

func Init(router *mux.Router, apps *apps.Service) {
	a := helloapp{
		mm:           apps.Mattermost,
		configurator: apps.Config,
	}

	subrouter := router.PathPrefix(constants.HelloAppPath).Subrouter()
	subrouter.HandleFunc("/mattermost-app.json", a.handleManifest).Methods("GET")
	subrouter.PathPrefix("/oauth2").HandlerFunc(a.handleOAuth).Methods(http.MethodGet, http.MethodPost)
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
			CallbackURL: "http://localhost:8065/plugins/apps/helloapp/oauth",
			Homepage:    "http://localhost:8065/plugins/apps/helloapp",
		})
}

func (h *helloapp) handleOAuth(w http.ResponseWriter, req *http.Request) {
	if h.OAuther == nil {
		http.Error(w, "OAuth not initialized", http.StatusInternalServerError)
		return
	}

	h.OAuther.ServeHTTP(w, req)
}
