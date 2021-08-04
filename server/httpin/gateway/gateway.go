package gateway

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type gateway struct {
	conf  config.Service
	proxy proxy.Service
}

func Init(router *mux.Router, conf config.Service, proxy proxy.Service, _ appservices.Service) {
	mm := conf.MattermostAPI()
	g := &gateway{
		conf:  conf,
		proxy: proxy,
	}

	subrouter := router.PathPrefix(config.PathApps).Subrouter()

	// Static
	subrouter.HandleFunc("/{appid}/"+apps.StaticFolder+"/{name}",
		httputils.CheckAuthorized(mm, g.static)).Methods(http.MethodGet)

	// Incoming remote webhooks
	subrouter.HandleFunc("/{appid}"+apps.PathWebhook+"/{path}",
		g.handleWebhook).Methods(http.MethodPost)

	// Remote OAuth2
	subrouter.HandleFunc("/{appid}"+config.PathRemoteOAuth2Connect,
		httputils.CheckAuthorized(mm, g.remoteOAuth2Connect)).Methods(http.MethodGet)
	subrouter.HandleFunc("/{appid}"+config.PathRemoteOAuth2Complete,
		httputils.CheckAuthorized(mm, g.remoteOAuth2Complete)).Methods(http.MethodGet)
}

func appIDVar(r *http.Request) apps.AppID {
	s, ok := mux.Vars(r)["appid"]
	if ok {
		return apps.AppID(s)
	}
	return ""
}
