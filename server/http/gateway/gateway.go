package gateway

import (
	"net/http"

	"github.com/gorilla/mux"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

type gateway struct {
	conf  config.Service
	proxy proxy.Service
	mm    *pluginapi.Client
}

func Init(router *mux.Router, mm *pluginapi.Client, conf config.Service, proxy proxy.Service, _ appservices.Service) {
	g := &gateway{
		conf:  conf,
		mm:    mm,
		proxy: proxy,
	}

	subrouter := router.PathPrefix(config.AppsPath).Subrouter()

	// Static
	subrouter.HandleFunc("/{app_id}/"+apps.StaticAssetsFolder+"/{name}",
		httputils.CheckAuthorized(mm, g.static)).Methods(http.MethodGet)

	// OAuth2
	subrouter.HandleFunc("/{app_id}"+apps.PathOAuthRedirect,
		httputils.CheckAuthorized(mm, g.remoteOAuth2Redirect)).Methods(http.MethodGet)
	subrouter.HandleFunc("/{app_id}"+apps.PathOAuthComplete,
		httputils.CheckAuthorized(mm, g.remoteOAuth2Complete)).Methods(http.MethodGet)

	// Webhooks
	subrouter.HandleFunc("/{app_id}"+config.Webhook+"/{path}", g.handleWebhook).Methods(http.MethodPost)
}
