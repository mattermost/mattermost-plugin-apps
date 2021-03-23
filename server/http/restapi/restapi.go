package restapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

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
	subrouter.HandleFunc(apps.DefaultBindingsCallPath, checkAuthorized(a.handleGetBindings)).Methods("GET")
	subrouter.HandleFunc(config.CallPath, a.handleCall).Methods("POST")
	subrouter.HandleFunc(config.SubscribePath, a.handleSubscribe).Methods("POST")
	subrouter.HandleFunc(config.UnsubscribePath, a.handleUnsubscribe).Methods("POST")

	subrouter.HandleFunc(config.KVPath+"/{key}", a.handleKV(a.kvGet)).Methods("GET")
	subrouter.HandleFunc(config.KVPath+"/{key}", a.handleKV(a.kvPut)).Methods("PUT", "POST")
	subrouter.HandleFunc(config.KVPath+"/", a.handleKV(a.kvList)).Methods("GET")
	subrouter.HandleFunc(config.KVPath+"/{key}", a.handleKV(a.kvHead)).Methods("HEAD")
	subrouter.HandleFunc(config.KVPath+"/{key}", a.handleKV(a.kvDelete)).Methods("DELETE")

	subrouter.HandleFunc(config.PathMarketplace, checkAuthorized(a.handleGetMarketplace)).Methods(http.MethodGet)
	subrouter.HandleFunc(config.StaticAssetPath+"/{app_id}/{name}", checkAuthorized(a.handleGetStaticAsset)).Methods(http.MethodGet)
	subrouter.HandleFunc(config.Webhook+"/{app_id}/{name}", a.handleWebhook).Methods(http.MethodPost)
}

func checkAuthorized(f func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		actingUserID := req.Header.Get("Mattermost-User-Id")
		if actingUserID == "" {
			httputils.WriteUnauthorizedError(w, errors.New("not authorized"))
			return
		}

		f(w, req, actingUserID)
	}
}
