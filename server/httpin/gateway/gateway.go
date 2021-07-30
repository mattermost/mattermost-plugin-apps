package gateway

import (
	"net/http"

	"github.com/gorilla/mux"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type gateway struct {
	mm    *pluginapi.Client
	log   utils.Logger
	conf  config.Service
	proxy proxy.Service
}

func Init(router *mux.Router, mm *pluginapi.Client, log utils.Logger, conf config.Service, p proxy.Service, _ appservices.Service) {
	g := &gateway{
		mm:    mm,
		log:   log,
		conf:  conf,
		proxy: p,
	}

	subrouter := router.PathPrefix(config.PathApps).Subrouter()

	// Static
	subrouter.HandleFunc("/{appid}/"+apps.StaticFolder+"/{name}",
		proxy.RequireUser(g.static)).Methods(http.MethodGet)

	// Incoming remote webhooks
	subrouter.HandleFunc("/{appid}"+apps.PathWebhook+"/{path}",
		g.handleWebhook).Methods(http.MethodPost)

	// Remote OAuth2
	subrouter.HandleFunc("/{appid}"+config.PathRemoteOAuth2Connect,
		proxy.RequireUser(g.remoteOAuth2Connect)).Methods(http.MethodGet)
	subrouter.HandleFunc("/{appid}"+config.PathRemoteOAuth2Complete,
		proxy.RequireUser(g.remoteOAuth2Complete)).Methods(http.MethodGet)
}

func appIDVar(r *http.Request) apps.AppID {
	s, ok := mux.Vars(r)["appid"]
	if ok {
		return apps.AppID(s)
	}
	return ""
}
