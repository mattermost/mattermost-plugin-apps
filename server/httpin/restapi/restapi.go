package restapi

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

type restapi struct {
	conf        config.Service
	proxy       proxy.Service
	appServices appservices.Service
}

func Init(router *mux.Router, conf config.Service, p proxy.Service, appServices appservices.Service) {
	mm := conf.MattermostAPI()
	a := &restapi{
		conf:        conf,
		proxy:       p,
		appServices: appServices,
	}

	subrouter := router.PathPrefix(appclient.PathAPI).Subrouter()

	// Proxy API, intended to be used by the user-agents (mobile, desktop, and
	// web).

	// Call
	subrouter.HandleFunc(config.PathCall, proxy.RequireUser(a.handleCall)).Methods("POST")
	// GetBindings
	subrouter.HandleFunc(apps.DefaultBindings.Path, proxy.RequireUser(a.handleGetBindings)).Methods("GET")
	// GetBotIDs
	subrouter.HandleFunc(appclient.PathBotIDs, proxy.RequireUser(a.handleGetBotIDs)).Methods("GET")
	// GetOAuthAppIDs
	subrouter.HandleFunc(appclient.PathOAuthAppIDs, proxy.RequireUser(a.handleGetOAuthAppIDs)).Methods("GET")

	// App Service API, intended to be used by Apps. Subscriptions, KV, OAuth2
	// services.
	subrouter.HandleFunc(appclient.PathSubscribe, proxy.RequireSysadmin(mm, a.handleSubscribe)).Methods("POST")
	subrouter.HandleFunc(appclient.PathSubscribe, proxy.RequireSysadmin(mm, a.handleGetSubscriptions)).Methods("GET")
	subrouter.HandleFunc(appclient.PathUnsubscribe, proxy.RequireSysadmin(mm, a.handleUnsubscribe)).Methods("POST")
	subrouter.HandleFunc(appclient.PathKV+"/{prefix}/{key}", proxy.RequireUser(a.kvGet)).Methods("GET")
	subrouter.HandleFunc(appclient.PathKV+"/{key}", proxy.RequireUser(a.kvGet)).Methods("GET")
	subrouter.HandleFunc(appclient.PathKV+"/{prefix}/{key}", proxy.RequireUser(a.kvPut)).Methods("PUT", "POST")
	subrouter.HandleFunc(appclient.PathKV+"/{key}", proxy.RequireUser(a.kvPut)).Methods("PUT", "POST")
	subrouter.HandleFunc(appclient.PathKV+"/{key}", proxy.RequireUser(a.kvDelete)).Methods("DELETE")
	subrouter.HandleFunc(appclient.PathKV+"/{prefix}/{key}", proxy.RequireUser(a.kvDelete)).Methods("DELETE")
	// TODO appid should come from OAuth2 user session, see
	// https://mattermost.atlassian.net/browse/MM-34377
	subrouter.HandleFunc(appclient.PathOAuth2App+"/{appid}", proxy.RequireUser(a.oauth2StoreApp)).Methods("PUT", "POST")
	subrouter.HandleFunc(appclient.PathOAuth2User+"/{appid}", proxy.RequireUser(a.oauth2StoreUser)).Methods("PUT", "POST")
	subrouter.HandleFunc(appclient.PathOAuth2User+"/{appid}", proxy.RequireUser(a.oauth2GetUser)).Methods("GET")

	// Admin API, can be used by plugins or by external services.
	subrouter.HandleFunc(config.PathMarketplace,
		proxy.RequireUser(a.handleGetMarketplace)).Methods(http.MethodGet)

	appsRouters := subrouter.PathPrefix(appclient.PathApps).Subrouter()
	appsRouters.HandleFunc("",
		proxy.RequireSysadminOrPlugin(mm, a.handleInstallApp)).Methods("POST")

	appRouter := appsRouters.PathPrefix(`/{appid:[A-Za-z0-9-_.]+}`).Subrouter()
	appRouter.HandleFunc("",
		proxy.RequireSysadminOrPlugin(mm, a.handleGetApp)).Methods("GET")
	appRouter.HandleFunc(appclient.PathEnable,
		proxy.RequireSysadminOrPlugin(mm, a.handleEnableApp)).Methods("POST")
	appRouter.HandleFunc(appclient.PathDisable,
		proxy.RequireSysadminOrPlugin(mm, a.handleDisableApp)).Methods("POST")
	appRouter.HandleFunc("",
		proxy.RequireSysadminOrPlugin(mm, a.handleUninstallApp)).Methods("DELETE")
}

func appIDVar(r *http.Request) apps.AppID {
	s, ok := mux.Vars(r)["appid"]
	if ok {
		return apps.AppID(s)
	}
	return ""
}
