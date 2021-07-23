package restapi

import (
	"net/http"

	"github.com/gorilla/mux"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type restapi struct {
	mm          *pluginapi.Client
	conf        config.Service
	proxy       proxy.Service
	appServices appservices.Service
	i18n        *i18n.Bundle
}

func Init(router *mux.Router, mm *pluginapi.Client, conf config.Service, proxy proxy.Service, appServices appservices.Service, i18nBundle *i18n.Bundle) {
	a := &restapi{
		mm:          mm,
		conf:        conf,
		proxy:       proxy,
		appServices: appServices,
		i18n:        i18nBundle,
	}

	subrouter := router.PathPrefix(mmclient.PathAPI).Subrouter()

	subrouter.HandleFunc(apps.DefaultBindings.Path,
		httputils.CheckAuthorized(mm, a.handleGetBindings)).Methods("GET")

	subrouter.HandleFunc(config.PathCall,
		httputils.CheckAuthorized(mm, a.handleCall)).Methods("POST")

	subrouter.HandleFunc(mmclient.PathSubscribe, a.handleSubscribe).Methods("POST")
	subrouter.HandleFunc(mmclient.PathUnsubscribe, a.handleUnsubscribe).Methods("POST")

	// Bot and OAuthApps checks
	subrouter.HandleFunc(mmclient.PathBotIDs,
		httputils.CheckAuthorized(mm, a.handleGetBotIDs)).Methods("GET")
	subrouter.HandleFunc(mmclient.PathOAuthAppIDs,
		httputils.CheckAuthorized(mm, a.handleGetOAuthAppIDs)).Methods("GET")

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

	appsRouters := subrouter.PathPrefix(mmclient.PathApps).Subrouter()
	appsRouters.HandleFunc("", httputils.CheckPluginIDOrUserSession(a.handleInstallApp)).Methods("POST")

	appRouter := appsRouters.PathPrefix(`/{appid:[A-Za-z0-9-_.]+}`).Subrouter()
	appRouter.HandleFunc("", httputils.CheckPluginIDOrUserSession(a.handleGetApp)).Methods("GET")
	appRouter.HandleFunc(mmclient.PathEnable, httputils.CheckPluginIDOrUserSession(a.handleEnableApp)).Methods("POST")
	appRouter.HandleFunc(mmclient.PathDisable, httputils.CheckPluginIDOrUserSession(a.handleDisableApp)).Methods("POST")
	appRouter.HandleFunc(mmclient.PathUninstall, httputils.CheckPluginIDOrUserSession(a.handleUninstallApp)).Methods("DELETE")
}

func actingID(r *http.Request) string {
	return r.Header.Get("Mattermost-User-Id")
}

func appIDVar(r *http.Request) apps.AppID {
	s, ok := mux.Vars(r)["appid"]
	if ok {
		return apps.AppID(s)
	}
	return ""
}
