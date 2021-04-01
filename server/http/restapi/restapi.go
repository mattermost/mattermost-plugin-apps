package restapi

import (
	"net/http"

	"github.com/gorilla/mux"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

type restapi struct {
	mm          *pluginapi.Client
	conf        config.Service
	proxy       proxy.Service
	appServices appservices.Service
}

func Init(router *mux.Router, mm *pluginapi.Client, conf config.Service, proxy proxy.Service, appServices appservices.Service) {
	a := &restapi{
		mm:          mm,
		conf:        conf,
		proxy:       proxy,
		appServices: appServices,
	}

	subrouter := router.PathPrefix(mmclient.PathAPI).Subrouter()

	subrouter.HandleFunc(apps.DefaultBindings.Path,
		httputils.CheckAuthorized(mm, a.handleGetBindings)).Methods("GET")

	subrouter.HandleFunc(config.PathCall, a.handleCall).Methods("POST")
	subrouter.HandleFunc(mmclient.PathSubscribe, a.handleSubscribe).Methods("POST")
	subrouter.HandleFunc(mmclient.PathUnsubscribe, a.handleUnsubscribe).Methods("POST")

	// KV APIs
	subrouter.HandleFunc(mmclient.PathKV+"/{prefix}/{key}", a.kvGet).Methods("GET")
	subrouter.HandleFunc(mmclient.PathKV+"/{key}", a.kvGet).Methods("GET")
	subrouter.HandleFunc(mmclient.PathKV+"/{prefix}/{key}", a.kvPut).Methods("PUT", "POST")
	subrouter.HandleFunc(mmclient.PathKV+"/{key}", a.kvPut).Methods("PUT", "POST")

	subrouter.HandleFunc(mmclient.PathKV+"/{key}", a.kvDelete).Methods("DELETE")
	subrouter.HandleFunc(mmclient.PathKV+"/{prefix}/{key}", a.kvDelete).Methods("DELETE")

	// TODO appid should come from OAuth2 user session, see
	// https://mattermost.atlassian.net/browse/MM-34377
	subrouter.HandleFunc(mmclient.PathOAuth2App+"/{appid}", a.oauth2StoreApp).Methods("PUT", "POST")
	subrouter.HandleFunc(mmclient.PathOAuth2User+"/{appid}", a.oauth2StoreUser).Methods("PUT", "POST")
	subrouter.HandleFunc(mmclient.PathOAuth2User+"/{appid}", a.oauth2GetUser).Methods("GET")

	subrouter.HandleFunc(config.PathMarketplace,
		httputils.CheckAuthorized(mm, a.handleGetMarketplace)).Methods(http.MethodGet)
}

func actingID(r *http.Request) string {
	return r.Header.Get("Mattermost-User-Id")
}

func sessionID(r *http.Request) string {
	return r.Header.Get("MM_SESSION_ID")
}

func appIDVar(r *http.Request) apps.AppID {
	s, ok := mux.Vars(r)["appid"]
	if ok {
		return apps.AppID(s)
	}
	return ""
}
