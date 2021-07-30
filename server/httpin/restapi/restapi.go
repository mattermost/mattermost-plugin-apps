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
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type restapi struct {
	mm          *pluginapi.Client
	log         utils.Logger
	conf        config.Service
	proxy       proxy.Service
	appServices appservices.Service
}

func Init(router *mux.Router, mm *pluginapi.Client, log utils.Logger, conf config.Service, proxy proxy.Service, appServices appservices.Service) {
	a := &restapi{
		mm:          mm,
		log:         log,
		conf:        conf,
		proxy:       proxy,
		appServices: appServices,
	}

	subrouter := router.PathPrefix(mmclient.PathAPI).Subrouter()

	// Proxy API, intended to be used by the user-agents (mobile, desktop, and
	// web).
	subrouter.HandleFunc(apps.DefaultBindings.Path,
		httputils.RequireUser(mm, a.handleGetBindings)).Methods("GET")
	subrouter.HandleFunc(config.PathCall,
		httputils.RequireUser(mm, a.handleCall)).Methods("POST")
	subrouter.HandleFunc(mmclient.PathBotIDs,
		httputils.CheckAuthorized(mm, a.handleGetBotIDs)).Methods("GET")
	subrouter.HandleFunc(mmclient.PathOAuthAppIDs,
		httputils.CheckAuthorized(mm, a.handleGetOAuthAppIDs)).Methods("GET")

	// App Service API, intended to be used by Apps. Subscriptions, KV, OAuth2
	// services.
	subrouter.HandleFunc(mmclient.PathSubscribe, a.handleSubscribe).Methods("POST")
	subrouter.HandleFunc(mmclient.PathUnsubscribe, a.handleUnsubscribe).Methods("POST")
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

	// Admin API, can be used by plugins or by external services.
	subrouter.HandleFunc(config.PathMarketplace,
		httputils.CheckAuthorized(mm, a.handleGetMarketplace)).Methods(http.MethodGet)

	appsRouters := subrouter.PathPrefix(mmclient.PathApps).Subrouter()
	appsRouters.HandleFunc("", httputils.CheckPluginIDOrUserSession(a.handleInstallApp)).Methods("POST")

	appRouter := appsRouters.PathPrefix(`/{appid:[A-Za-z0-9-_.]+}`).Subrouter()
	appRouter.HandleFunc("", httputils.RequireUserOrPlugin(a.handleGetApp)).Methods("GET")
	appRouter.HandleFunc(mmclient.PathEnable, httputils.CheckPluginIDOrUserSession(a.handleEnableApp)).Methods("POST")
	appRouter.HandleFunc(mmclient.PathDisable, httputils.CheckPluginIDOrUserSession(a.handleDisableApp)).Methods("POST")
	appRouter.HandleFunc("", httputils.CheckPluginIDOrUserSession(a.handleUninstallApp)).Methods("DELETE")
}

func appIDVar(r *http.Request) apps.AppID {
	s, ok := mux.Vars(r)["appid"]
	if ok {
		return apps.AppID(s)
	}
	return ""
}
