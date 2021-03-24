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

	subrouter := router.PathPrefix(config.AppsPath).Subrouter()

	subrouter.HandleFunc("/{app_id}/"+apps.StaticAssetsFolder+"/{name}",
		httputils.CheckAuthorized(mm, p.handleGetStaticAsset)).Methods(http.MethodGet)

	subrouter.HandleFunc("/{app_id}"+apps.PathOAuthRedirect,
		httputils.CheckAuthorized(mm, p.handleGetOAuth2RemoteRedirect)).Methods(http.MethodGet)
}
