package restapi

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
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

	api := router.PathPrefix(path.API).Subrouter()

	// Proxy API, intended to be used by the user-agents (mobile, desktop, and
	// web).
	a.initCall(api)

	// User-agent APIs.
	a.initGetBindings(api)
	a.initGetBotIDs(api)
	a.initGetOAuthAppIDs(api)

	// App Service API, intended to be used by Apps. Subscriptions, KV, OAuth2
	// services.
	a.initSubscriptions(api, mm)
	a.initKV(api)
	a.initOAuth2Store(api)

	// Admin API, can be used by plugins, external services, or the user agent.
	a.initAdmin(api, mm)
	a.initGetApp(api, mm)
	a.initMarketplace(api)
}

func appIDVar(r *http.Request) apps.AppID {
	s, ok := mux.Vars(r)["appid"]
	if ok {
		return apps.AppID(s)
	}
	return ""
}
