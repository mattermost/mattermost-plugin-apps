package restapi

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

type restapi struct {
	conf        config.Service
	proxy       proxy.Service
	appServices appservices.Service
}

func Init(rh *httpin.Handler, conf config.Service, p proxy.Service, appServices appservices.Service) {
	a := &restapi{
		conf:        conf,
		proxy:       p,
		appServices: appServices,
	}

	rh = rh.PathPrefix(path.API)

	a.initPing(rh)

	// Proxy API, intended to be used by the user-agents (mobile, desktop, and
	// web).
	a.initCall(rh)

	// User-agent APIs.
	a.initGetBindings(rh)
	a.initGetBotIDs(rh)
	a.initGetOAuthAppIDs(rh)

	// App Service API, intended to be used by Apps. Subscriptions, KV, OAuth2
	// services.
	a.initSubscriptions(rh)
	a.initKV(rh)
	a.initOAuth2Store(rh)

	// Admin API, can be used by plugins, external services, or the user agent.
	a.initAdmin(rh)
	a.initGetApp(rh)
	a.initMarketplace(rh)
}

func appIDVar(r *http.Request) apps.AppID {
	s, ok := mux.Vars(r)["appid"]
	if ok {
		return apps.AppID(s)
	}
	return ""
}
