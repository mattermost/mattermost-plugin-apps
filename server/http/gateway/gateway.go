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
	p := &gateway{
		conf:  conf,
		mm:    mm,
		proxy: proxy,
	}

	subrouter := router.PathPrefix(config.PathApps).Subrouter()

	// Static
	subrouter.HandleFunc("/{app_id}/"+apps.StaticFolder+"/{name}",
		httputils.CheckAuthorized(mm, p.static)).Methods(http.MethodGet)

	// Remote OAuth2
	subrouter.HandleFunc("/{app_id}"+config.PathRemoteOAuth2Redirect,
		httputils.CheckAuthorized(mm, p.remoteOAuth2Redirect)).Methods(http.MethodGet)
	subrouter.HandleFunc("/{app_id}"+config.PathRemoteOAuth2Complete,
		httputils.CheckAuthorized(mm, p.remoteOAuth2Complete)).Methods(http.MethodGet)
}
