package helloapp

import (
	"net/http"

	"github.com/gorilla/mux"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/experimental/oauther"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

const AppSecret = "1234"

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
		configurator: apps.Configurator,
	}

	subrouter := router.PathPrefix(constants.HelloAppPath).Subrouter()
	subrouter.HandleFunc("/mattermost-app.json", a.handleManifest).Methods("GET")

	subrouter.HandleFunc("/wish/install", a.handleInstall).Methods("POST")
	subrouter.PathPrefix("/oauth2").HandlerFunc(a.handleOAuth).Methods(http.MethodGet, http.MethodPost)
}

func (h *helloapp) handleManifest(w http.ResponseWriter, req *http.Request) {
	conf := h.configurator.GetConfig()

	rootURL := conf.PluginURL + constants.HelloAppPath

	httputils.WriteJSON(w,
		apps.Manifest{
			AppID:       "hello",
			DisplayName: "Hallo სამყარო",
			Description: "Hallo სამყარო test app",
			RootURL:     rootURL,
			RequestedPermissions: []apps.PermissionType{
				apps.PermissionUserJoinedChannelNotification,
				apps.PermissionActAsUser,
				apps.PermissionActAsBot,
			},
			Install: &apps.Wish{
				URL: rootURL + "/wish/install",
			},
			CallbackURL: rootURL + "/oauth",
			Homepage:    rootURL,
		})
}
