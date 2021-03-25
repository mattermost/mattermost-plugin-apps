package restapi

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

	subrouter := router.PathPrefix(config.APIPath).Subrouter()

	subrouter.HandleFunc(apps.DefaultBindingsCallPath,
		httputils.CheckAuthorized(mm, a.handleGetBindings)).Methods("GET")

	subrouter.HandleFunc(config.CallPath, a.handleCall).Methods("POST")
	subrouter.HandleFunc(config.SubscribePath, a.handleSubscribe).Methods("POST")
	subrouter.HandleFunc(config.UnsubscribePath, a.handleUnsubscribe).Methods("POST")

	subrouter.HandleFunc(config.KVPath+"/{key}", a.handleKV(a.kvGet)).Methods("GET")
	subrouter.HandleFunc(config.KVPath+"/{key}", a.handleKV(a.kvPut)).Methods("PUT", "POST")
	subrouter.HandleFunc(config.KVPath+"/", a.handleKV(a.kvList)).Methods("GET")
	subrouter.HandleFunc(config.KVPath+"/{key}", a.handleKV(a.kvHead)).Methods("HEAD")
	subrouter.HandleFunc(config.KVPath+"/{key}", a.handleKV(a.kvDelete)).Methods("DELETE")
	subrouter.HandleFunc(config.Webhook+"/{app_id}/{name}", a.handleWebhook).Methods(http.MethodPost)
	subrouter.HandleFunc(config.PathMarketplace,
		httputils.CheckAuthorized(mm, a.handleGetMarketplace)).Methods(http.MethodGet)
}
